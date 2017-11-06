package failover_test

import (
	"fmt"
	"testing"

	"github.com/spotahome/redis-operator/pkg/failover"
	"github.com/stretchr/testify/assert"
)

func TestSetReadyConditionNew(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}

	status.SetReadyCondition()

	assert.Len(status.Conditions, 1, "Condition not Saved")
	assert.Equal(failover.ConditionReady, status.Conditions[0].Type, "Condition incorrect")
}

func TestSetReadyConditionConsecutive(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}

	status.SetReadyCondition()
	status.SetReadyCondition()

	assert.Len(status.Conditions, 1, "Condition saved twice")
}

func TestSetReadyConditionNoConsecutive(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}

	status.SetReadyCondition()
	status.SetNotReadyCondition()
	status.SetReadyCondition()

	assert.Len(status.Conditions, 3, "Condition not saved")
}

func TestSetNotReadyConditionNew(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}

	status.SetNotReadyCondition()

	assert.Len(status.Conditions, 1, "Condition not Saved")
	assert.Equal(failover.ConditionNotReady, status.Conditions[0].Type, "Condition incorrect")
}

func TestSetNotReadyConditionConsecutive(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}

	status.SetNotReadyCondition()
	status.SetNotReadyCondition()

	assert.Len(status.Conditions, 1, "Condition saved twice")
}

func TestSetNotReadyConditionNoConsecutive(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}

	status.SetNotReadyCondition()
	status.SetReadyCondition()
	status.SetNotReadyCondition()

	assert.Len(status.Conditions, 3, "Condition not saved")
}

func TestAppendUpdatingRedisCondition(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}
	reason := "reason"

	status.AppendUpdatingRedisCondition(reason)

	assert.Len(status.Conditions, 1, "Condition not saved")
	assert.Equal(failover.ConditionUpdatingRedis, status.Conditions[0].Type, "Condition incorrect")
	assert.Equal(reason, status.Conditions[0].Reason, "Reason not saved")
}

func TestAppendUpdatingSentinelCondition(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}
	reason := "reason"

	status.AppendUpdatingSentinelCondition(reason)

	assert.Len(status.Conditions, 1, "Condition not saved")
	assert.Equal(failover.ConditionUpdatingSentinel, status.Conditions[0].Type, "Condition incorrect")
	assert.Equal(reason, status.Conditions[0].Reason, "Reason not saved")
}

func TestAppendScalingRedisUpCondition(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}

	status.AppendScalingRedisUpCondition(3, 4)

	assert.Len(status.Conditions, 1, "Condition not saved")
	assert.Equal(failover.ConditionScalingRedisUp, status.Conditions[0].Type, "Condition incorrect")
}

func TestAppendScalingRedisDownCondition(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}

	status.AppendScalingRedisDownCondition(4, 3)

	assert.Len(status.Conditions, 1, "Condition not saved")
	assert.Equal(failover.ConditionScalingRedisDown, status.Conditions[0].Type, "Condition incorrect")
}

func TestAppendScalingSentinelUpCondition(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}

	status.AppendScalingSentinelUpCondition(3, 4)

	assert.Len(status.Conditions, 1, "Condition not saved")
	assert.Equal(failover.ConditionScalingSentinelUp, status.Conditions[0].Type, "Condition incorrect")
}

func TestAppendScalingSentinelDownCondition(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}

	status.AppendScalingSentinelDownCondition(4, 3)

	assert.Len(status.Conditions, 1, "Condition not saved")
	assert.Equal(failover.ConditionScalingSentinelDown, status.Conditions[0].Type, "Condition incorrect")
}

func TestSetMaster(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}
	master := "0.0.0.0"

	status.SetMaster(master)

	assert.Equal(status.Master, master, "Master not saved")
}

func TestConditionLength(t *testing.T) {
	assert := assert.New(t)
	status := failover.RedisFailoverStatus{}
	reason := "reason"

	for i := 0; i < 10; i++ {
		status.AppendUpdatingRedisCondition(reason)
	}

	status.AppendUpdatingSentinelCondition(reason)

	assert.Len(status.Conditions, 10, "Conditions arrays bigger than expected")
	for i := 0; i < 9; i++ {
		assert.Equal(failover.ConditionUpdatingRedis, status.Conditions[i].Type, fmt.Sprintf("Condition %d different than %s", i, failover.ConditionUpdatingRedis))
	}
	assert.Equal(failover.ConditionUpdatingSentinel, status.Conditions[9].Type, fmt.Sprintf("Condition 9 different than %s", failover.ConditionUpdatingSentinel))
}
