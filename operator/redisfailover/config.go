package redisfailover

// Config is the configuration for the redis operator.
type Config struct {
	// Optional labels that can be added appart from the default ones.
	Labels map[string]string

	ListenAddress string
	MetricsPath   string
}
