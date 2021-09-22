package service

import (
	"fmt"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
)

// GetRedisShutdownConfigMapName returns the name for redis configmap
func GetRedisShutdownConfigMapName(rf *redisfailoverv1.RedisFailover) string {
	if rf.Spec.Redis.ShutdownConfigMap != "" {
		return rf.Spec.Redis.ShutdownConfigMap
	}
	return GetRedisShutdownName(rf)
}

// GetRedisNameByRole returns the name for redis master or slave resources
func GetRedisNameByRole(rf *redisfailoverv1.RedisFailover, role string) string {
	return generateName(redisName, rf.Name+"-"+role)
}

// GetRedisName returns the name for redis resources
func GetRedisName(rf *redisfailoverv1.RedisFailover) string {
	return generateName(redisName, rf.Name)
}

// GetRedisShutdownName returns the name for redis resources
func GetRedisShutdownName(rf *redisfailoverv1.RedisFailover) string {
	return generateName(redisShutdownName, rf.Name)
}

// GetRedisReadinessName returns the name for redis resources
func GetRedisReadinessName(rf *redisfailoverv1.RedisFailover) string {
	return generateName(redisReadinessName, rf.Name)
}

// GetSentinelName returns the name for sentinel resources
func GetSentinelName(rf *redisfailoverv1.RedisFailover) string {
	return generateName(sentinelName, rf.Name)
}

func generateName(typeName, metaName string) string {
	return fmt.Sprintf("%s%s-%s", baseName, typeName, metaName)
}
