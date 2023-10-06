package v1

const (
	defaultRedisNumber    = 3
	defaultSentinelNumber = 3
	defaultRedisPort      = 6379
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
	DefaultSentinelExporterImage = "quay.io/oliver006/redis_exporter:v1.43.0"
	DefaultExporterImage         = "quay.io/oliver006/redis_exporter:v1.43.0"
	DefaultImage                 = "redis:6.2.6-alpine"
)
