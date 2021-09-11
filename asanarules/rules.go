package asanarules

import "fmt"
import "time"

type periodic struct {
	duration time.Duration
	done     chan bool
}

var periodics = []*periodic{}

func Every(d time.Duration) *periodic {
	ret := &periodic{
		duration: d,
		done:     make(chan bool),
	}

	periodics = append(periodics, ret)

	return ret
}

func Loop() {
	for _, periodic := range periodics {
		periodic.start()
	}

	for _, periodic := range periodics {
		periodic.wait()
	}
}

func (p *periodic) MyTasks() *periodic {
	return p
}

func (p *periodic) start() {
	go p.loop()
}

func (p *periodic) wait() {
	<-p.done
}

func (p *periodic) loop() {
	ticker := time.NewTicker(p.duration)

	for {
		<-ticker.C
		p.exec()
	}

	close(p.done)
}

func (p *periodic) exec() {
	fmt.Printf("exec\n")
}
