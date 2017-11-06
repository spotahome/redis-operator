package config

const (
	// Kind defines the CRD inside k8s
	Kind = "RedisFailover"
	// APIName is the name that will be used inside the k8s API to access the CRD resources
	APIName = "redisfailovers"
	// Domain inside the TPR is registered
	Domain = "spotahome.com"
	// Version of the TPR
	Version = "v1alpha1"
)

const (
	// ExporterImage defines the redis exporter image
	ExporterImage = "oliver006/redis_exporter"
	// ExporterImageVersion defines the redis exporter version
	ExporterImageVersion = "v0.11.3"
	// RedisToolkitImage defines the redis toolkit image
	RedisToolkitImage = "quay.io/spotahome/redis-operator-toolkit"
	// RedisToolkitImageVersion defines the redis toolkit image version
	RedisToolkitImageVersion = "1.0.0"
	// RedisImage defines the redis image
	RedisImage = "redis"
	// RedisImageVersion defines the redis image version
	RedisImageVersion = "3.2-alpine"
)
