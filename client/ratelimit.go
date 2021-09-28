package client

import "net/http"
import "strconv"
import "sync"
import "time"

type RateLimit struct {
	perSecond float64
	balance   float64
	limit     float64
	updated   time.Time
	mu        sync.Mutex
}

func NewRateLimit(perSecond, limit float64) *RateLimit {
	return &RateLimit{
		perSecond: perSecond,
		balance:   limit,
		limit:     limit,
		updated:   time.Now(),
	}
}

func NewRateLimitPerMinute(perMinute, limit float64) *RateLimit {
	return NewRateLimit(perMinute/60, limit)
}

// Acquire sufficient rate quota to execute 1 operation
func (rl *RateLimit) Acquire1() {
	rl.AcquireN(1.0)
}

// Acquire sufficient rate quota to execute /cost/ operations
func (rl *RateLimit) AcquireN(cost float64) {
	for {
		rl.mu.Lock()

		rl.replenish()

		if rl.balance >= cost {
			rl.balance -= cost
			rl.mu.Unlock()
			return
		}

		costDelta := cost - rl.balance
		sleep := time.Duration(costDelta / rl.perSecond * float64(time.Second))
		rl.mu.Unlock()
		time.Sleep(sleep)
	}
}

func (rl *RateLimit) MaybeRetryAfter(resp *http.Response) error {
	header := resp.Header.Get("Retry-After")
	if header == "" {
		return nil
	}

	retryAfter, err := strconv.ParseInt(header, 10, 64)
	if err != nil {
		return err
	}

	rl.RetryAfter(retryAfter)
	return nil
}

func (rl *RateLimit) RetryAfter(seconds int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	target := 1.0 - (float64(seconds) * rl.perSecond)
	if target < rl.balance {
		rl.balance = target
	}
}

// Must be called with rl.mu already locked
func (rl *RateLimit) replenish() {
	now := time.Now()
	timeDelta := now.Sub(rl.updated)
	balanceDelta := timeDelta.Seconds() * rl.perSecond

	rl.balance += balanceDelta
	if rl.balance > rl.limit {
		rl.balance = rl.limit
	}

	rl.updated = now
}
