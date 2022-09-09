package v1

const (
	defaultRedisNumber           = 3
	defaultSentinelNumber        = 3
	defaultSentinelExporterImage = "quay.io/oliver006/redis_exporter:v1.43.0"
	defaultExporterImage         = "quay.io/oliver006/redis_exporter:v1.43.0"
	defaultImage                 = "redis:6.2.6-alpine"
	defaultRedisPort             = 6379
	defaultAdminUser             = "default"
	// AdminACL is a exported constant to be used in cases where Admin User is not "default"
	AdminACL = "allkeys -@all +client +ping +info +config|get +cluster|info +slowlog +latency +memory +select +get +scan +xinfo +type +pfcount +strlen +llen +scard +zcard +hlen +xlen +eval +@admin"
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
