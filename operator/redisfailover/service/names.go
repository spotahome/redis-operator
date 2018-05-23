package service

import (
	"fmt"

	redisfailoverv1alpha2 "github.com/spotahome/redis-operator/api/redisfailover/v1alpha2"
)

// GetRedisConfigMapName returns the name for redis configmap
func GetRedisConfigMapName(rf *redisfailoverv1alpha2.RedisFailover) string {
	if rf.Spec.Redis.ConfigMap != "" {
		return rf.Spec.Redis.ConfigMap
	}
	return GetRedisName(rf)
}

// GetRedisName returns the name for redis resources
func GetRedisName(rf *redisfailoverv1alpha2.RedisFailover) string {
	return generateName(redisName, rf.Name)
}

// GetSentinelConfigMapName returns the name for sentinel configmap
func GetSentinelConfigMapName(rf *redisfailoverv1alpha2.RedisFailover) string {
	if rf.Spec.Sentinel.ConfigMap != "" {
		return rf.Spec.Sentinel.ConfigMap
	}
	return GetSentinelName(rf)
}

// GetSentinelName returns the name for sentinel resources
func GetSentinelName(rf *redisfailoverv1alpha2.RedisFailover) string {
	return generateName(sentinelName, rf.Name)
}

func generateName(typeName, metaName string) string {
	return fmt.Sprintf("%s%s-%s", baseName, typeName, metaName)
}
