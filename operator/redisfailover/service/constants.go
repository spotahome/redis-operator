package service

import (
	"fmt"
	"os"
)

const (
	logNameField      = "redisfailover"
	logNamespaceField = "namespace"
)


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
	//appLabel               = "redis-failover"
	hostnameTopologyKey    = "kubernetes.io/hostname"
)

var (
	exporterImage = fmt.Sprintf("%s:%s", ExporterImage, ExporterImageVersion)

	// ExporterImage defines the redis exporter image
	ExporterImage = os.Getenv("REDIS_EXPORTER_IMAGE")
	// ExporterImageVersion defines the redis exporter version
	ExporterImageVersion = os.Getenv("REDIS_EXPORTER_IMAGE_VERSION")
	// RedisImage defines the redis image
	RedisImage = os.Getenv("REDIS_IMAGE")
	// RedisImageVersion defines the redis image version
	RedisImageVersion = os.Getenv("REDIS_IMAGE_VERSION")

  //app label variable to have more redis operators in cluster
	//Manage a Redis Failover deployment
  appLabel = os.Getenv("APP_LABEL")
// variables refering to the redis exporter port
)
