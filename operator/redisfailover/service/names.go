package service

import (
	"fmt"

	redisfailoverv1alpha2 "github.com/spotahome/redis-operator/api/redisfailover/v1alpha2"
)

const (
	maxNameLength = 60
)

// GetRedisShutdownConfigMapName returns the name for redis configmap
func GetRedisShutdownConfigMapName(rf *redisfailoverv1alpha2.RedisFailover) string {
	if rf.Spec.Redis.ShutdownConfigMap != "" {
		return rf.Spec.Redis.ShutdownConfigMap
	}
	return GetRedisShutdownName(rf)
}

// GetRedisName returns the name for redis resources
func GetRedisName(rf *redisfailoverv1alpha2.RedisFailover) string {
	return generateName(redisName, rf.Name)
}

// GetRedisShutdownName returns the name for redis resources
func GetRedisShutdownName(rf *redisfailoverv1alpha2.RedisFailover) string {
	return generateName(redisShutdownName, rf.Name)
}

// GetSentinelName returns the name for sentinel resources
func GetSentinelName(rf *redisfailoverv1alpha2.RedisFailover) string {
	return generateName(sentinelName, rf.Name)
}

func generateName(typeName, metaName string) string {
	generatedName := fmt.Sprintf("%s%s-%s", baseName, typeName, metaName)

	// If lenth higher than 60 characters, truncate the name so it does not cause problem when creating resources
	// Applicable for names and labels
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
	if len(generatedName) > maxNameLength {
		generatedName = generatedName[:maxNameLength]
	}

	return generatedName
}
