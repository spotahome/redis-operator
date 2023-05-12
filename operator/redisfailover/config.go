package redisfailover

// Config is the configuration for the redis operator.
type Config struct {
	ListenAddress   string
	MetricsPath     string
	Concurrency     int
	PDBMinAvailable string
}
