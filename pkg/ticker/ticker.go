// https://github.com/edward-zhu/go-ticker/blob/master/ticker.go

package ticker

import (
	"sync"
	"time"
)

// Ticker is an implementation of resetable ticker
//
// It keeps a constant Duration D, and when `Reset()`` is called,
// the ticker will be reset and next tick will arrive at least after
// D from the reset time.
type Ticker struct {
	Duration time.Duration
	C        <-chan time.Time
	c        chan time.Time
	resetCh  chan struct{}
	stopOnce sync.Once
	stopCh   chan struct{}
}

// NewTicker returns a new Ticker with duration d
func NewTicker(d time.Duration) *Ticker {
	c := make(chan time.Time)
	t := &Ticker{
		Duration: d,
		C:        c,
		c:        c,
		resetCh:  make(chan struct{}),
		stopCh:   make(chan struct{}),
	}

	go t.loop()

	return t
}

func (t *Ticker) loop() {
	afterCh := t.next()

	for {
		if afterCh == nil {
			return
		}

		select {
		case tick := <-afterCh:
			afterCh = t.send(tick)
		case <-t.resetCh:
			afterCh = t.next()
		case <-t.stopCh:
			close(t.c)
			t.c = nil
			return
		}
	}
}

func (t *Ticker) next() <-chan time.Time {
	return time.After(t.Duration)
}

func (t *Ticker) send(tick time.Time) <-chan time.Time {
	select {
	case t.c <- tick:
	case <-t.stopCh:
		return nil
	}

	return t.next()
}

// Reset will reset the ticker, and next tick will arrive after
// d from the reset time.
func (t *Ticker) Reset() {
	t.resetCh <- struct{}{}
}

// Stop turns off the ticker. There will be no more tick arrives
// after Stop() is called.
func (t *Ticker) Stop() {
	t.stopOnce.Do(func() { close(t.stopCh) })
}
