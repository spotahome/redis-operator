package v1

const (
	defaultRedisNumber           = 3
	defaultSentinelNumber        = 3
	defaultSentinelExporterImage = "leominov/redis_sentinel_exporter:1.3.0"
	defaultExporterImage         = "oliver006/redis_exporter:v1.3.5-alpine"
	defaultImage                 = "redis:5.0-alpine"
	defaultRedisPort             = "6379"
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
