package operator

import (
	"time"
)

// Config is the controller configuration.
type Config struct {
	// ResyncPeriod is the resync period of the operator.
	ResyncPeriod time.Duration
}
