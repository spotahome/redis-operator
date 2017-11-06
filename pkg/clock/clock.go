package clock

import (
	"time"
)

// Clock is an interface that will manage the time.
type Clock interface {
	After(d time.Duration) <-chan time.Time
	Now() time.Time
	Sleep(d time.Duration)
	Tick(d time.Duration) <-chan time.Time
	NewTicker(d time.Duration) *time.Ticker
	NewTimer(d time.Duration) *time.Timer
}

// New returns a new clock.
func New() Clock {
	return &clock{}
}

// clock is a real clock implementation of Clock interface.
type clock struct{}

func (c *clock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
func (c *clock) Now() time.Time {
	return time.Now()
}
func (c *clock) Sleep(d time.Duration) {
	time.Sleep(d)
}
func (c *clock) Tick(d time.Duration) <-chan time.Time {
	return time.Tick(d)
}
func (c *clock) NewTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}
func (c *clock) NewTimer(d time.Duration) *time.Timer {
	return time.NewTimer(d)
}

var base = &clock{}

// Base returns the base clock.
func Base() Clock {
	return base
}

// After returns a channel that will send a message when the duration passed reaches.
func After(d time.Duration) <-chan time.Time {
	return base.After(d)
}

// Now returns the current time.
func Now() time.Time {
	return base.Now()
}

// Sleep will sleep until the duration is reached.
func Sleep(d time.Duration) {
	base.Sleep(d)
}

// Tick returns a new channel that returns at a constant pace of duration.
func Tick(d time.Duration) <-chan time.Time {
	return base.Tick(d)
}

// NewTicker returns a new ticker.
func NewTicker(d time.Duration) *time.Ticker {
	return base.NewTicker(d)
}

// NewTimer returns a new timer.
func NewTimer(d time.Duration) *time.Timer {
	return base.NewTimer(d)
}
