package controller

import (
	"time"
)

// Config is the controller configuration.
type Config struct {
	ResyncPeriod time.Duration
	Namespace    string
}
