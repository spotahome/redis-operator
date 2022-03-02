package util

import (
	"time"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
)

// RedisNeedsRestart returns whether the Redis pods need to be restarted
func RedisNeedsRestart(rf *redisfailoverv1.RedisFailover) bool {
	now := time.Now().UTC()
	if rf.Spec.Redis.RestartAt == nil {
		return false
	}
	if rf.Spec.Redis.RestartAt.Equal(rf.Status.RedisRestartedAt) {
		return false
	}
	return now.After(rf.Spec.Redis.RestartAt.Time)
}

// SentinelNeedsRestart retuns whether the Sentinel pods need to be restarted
func SentinelNeedsRestart(rf *redisfailoverv1.RedisFailover) bool {
	now := time.Now().UTC()
	if rf.Spec.Sentinel.RestartAt == nil {
		return false
	}
	if rf.Spec.Sentinel.RestartAt.Equal(rf.Status.SentinelRestartedAt) {
		return false
	}
	return now.After(rf.Spec.Sentinel.RestartAt.Time)
}
