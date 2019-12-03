package service

const (
	logNameField      = "redisfailover"
	logNamespaceField = "namespace"
)

// variables refering to the redis exporter port
const (
	exporterPort                 = 9121
	exporterPortName             = "http-metrics"
	exporterContainerName        = "redis-exporter"
	exporterDefaultRequestCPU    = "25m"
	exporterDefaultLimitCPU      = "50m"
	exporterDefaultRequestMemory = "50Mi"
	exporterDefaultLimitMemory   = "100Mi"
)

const (
	description            = "Manage a Redis Failover deployment"
	baseName               = "rf"
	bootstrapName          = "b"
	sentinelName           = "s"
	sentinelRoleName       = "sentinel"
	sentinelConfigFileName = "sentinel.conf"
	redisConfigFileName    = "redis.conf"
	redisName              = "r"
	redisShutdownName      = "r-s"
	redisReadinessName     = "r-readiness"
	redisRoleName          = "redis"
	redisGroupName         = "mymaster"
	appLabel               = "redis-failover"
	hostnameTopologyKey    = "kubernetes.io/hostname"
)
