package controller

import "time"

// Defaults.
const (
	defResyncInterval       = 30 * time.Second
	defConcurrentWorkers    = 1
	defProcessingJobRetries = 3
)

// Config is the controller configuration.
type Config struct {
	// name of the controller.
	Name string
	// ConcurrentWorkers is the number of concurrent workers the controller will have running processing events.
	ConcurrentWorkers int
	// ResyncInterval is the interval the controller will process all the selected resources.
	ResyncInterval time.Duration
	// ProcessingJobRetries is the number of times the job will try to reprocess the event before returning a real error.
	ProcessingJobRetries int
}

func (c *Config) setDefaults() {
	if c.ConcurrentWorkers <= 0 {
		c.ConcurrentWorkers = defConcurrentWorkers
	}

	if c.ResyncInterval <= 0 {
		c.ResyncInterval = defResyncInterval
	}

	if c.ProcessingJobRetries <= 0 {
		c.ProcessingJobRetries = defProcessingJobRetries
	}
}
