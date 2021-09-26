package client

import "context"

import "golang.org/x/sync/semaphore"

type ConcurrencyLimit struct {
	sem *semaphore.Weighted
}

func NewConcurrencyLimit(limit int64) *ConcurrencyLimit {
	return &ConcurrencyLimit{
		sem: semaphore.NewWeighted(limit),
	}
}

func (cl *ConcurrencyLimit) Acquire1() {
	cl.AcquireN(1)
}

func (cl *ConcurrencyLimit) AcquireN(cost int64) {
	err := cl.sem.Acquire(context.TODO(), cost)
	if err != nil {
		panic(err)
	}
}

func (cl *ConcurrencyLimit) Release1() {
	cl.ReleaseN(1)
}

func (cl *ConcurrencyLimit) ReleaseN(cost int64) {
	cl.sem.Release(cost)
}
