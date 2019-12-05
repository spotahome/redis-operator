package v1

const (
	defaultRedisNumber    = 3
	defaultSentinelNumber = 3
	defaultSentinelExporterImage = "leominov/redis_sentinel_exporter:1.3.0"
	defaultExporterImage  = "oliver006/redis_exporter:v0.33.0"
	defaultImage          = "redis:5.0-alpine"
)

var (
	defaultSentinelCustomConfig = []string{
		"down-after-milliseconds 5000",
		"failover-timeout 10000",
	}
)
