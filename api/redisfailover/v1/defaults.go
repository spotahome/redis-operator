package v1

const (
	defaultRedisNumber           = 3
	defaultSentinelNumber        = 3
	defaultSentinelExporterImage = "quay.io/oliver006/redis_exporter:v1.43.0"
	defaultExporterImage         = "quay.io/oliver006/redis_exporter:v1.43.0"
	defaultImage                 = "redis:6.2.6-alpine"
	defaultRedisPort             = 6379
	HealthyState                 = "Healthy"
	NotHealthyState              = "NotHealthy"
)

var (
	defaultSentinelCustomConfig = []string{
		"down-after-milliseconds 5000",
		"failover-timeout 10000",
	}
	defaultRedisCustomConfig = []string{
		"replica-priority 100",
	}
	bootstrappingRedisCustomConfig = []string{
		"replica-priority 0",
	}
)
