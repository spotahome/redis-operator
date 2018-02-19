package service

import (
	"fmt"
)

const (
	logNameField      = "redisfailover"
	logNamespaceField = "namespace"
)

const (
	// ExporterImage defines the redis exporter image
	ExporterImage = "oliver006/redis_exporter"
	// ExporterImageVersion defines the redis exporter version
	ExporterImageVersion = "v0.11.3"
	// RedisImage defines the redis image
	RedisImage = "redis"
	// RedisImageVersion defines the redis image version
	RedisImageVersion = "3.2-alpine"
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
	redisRoleName          = "redis"
	redisGroupName         = "mymaster"
	appLabel               = "redis-failover"
	hostnameTopologyKey    = "kubernetes.io/hostname"
)

var (
	exporterImage = fmt.Sprintf("%s:%s", ExporterImage, ExporterImageVersion)
)
