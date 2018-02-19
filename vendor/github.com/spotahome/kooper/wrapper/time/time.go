package time

import "time"

// Time is a simple wrapper over a subset of time package from the standard
// library. Very useful for testing (mocks).
type Time interface {
	After(d time.Duration) <-chan time.Time
	NewTicker(d time.Duration) *time.Ticker
}

// Base is the base time wrapper.
var Base = &wrapper{}

type wrapper struct{}

func (w *wrapper) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
func (w *wrapper) NewTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}
