package failover

import (
	"fmt"
	"time"
)

// Phase of the RF status
type Phase string

// ConditionType defines the condition that the RF can have
type ConditionType string

const (
	// PhaseNone is the initial phase, when the redisfailover is created
	PhaseNone Phase = ""
	// PhaseCreating defines that the redisfailover is being created, but no completed
	PhaseCreating Phase = "Creating"
	// PhaseRunning means that the redisfailover created and running
	PhaseRunning Phase = "Running"
	// PhaseFailed means that it's been a failure when creating the redisfailover
	PhaseFailed Phase = "Failed"
)

const (
	// ConditionReady defines that the application is ready for accepting connections
	ConditionReady ConditionType = "Ready"
	// ConditionNotReady defines that the application is NOT ready for accepting connections
	ConditionNotReady ConditionType = "NotReady"
	// ConditionRecovering defines that the application is becoming ready from a NotReady state
	ConditionRecovering ConditionType = "Recovering"
	// ConditionUpdatingRedis defines that the redis nodes are being updated
	ConditionUpdatingRedis ConditionType = "UpdatingRedis"
	// ConditionUpdatingSentinel defines that the sentinel nodes are being updated
	ConditionUpdatingSentinel ConditionType = "UpdatingSentinel"
	// ConditionScalingRedisUp defines that the redis nodes are being increased
	ConditionScalingRedisUp ConditionType = "ScalingRedisUp"
	// ConditionScalingRedisDown defines that the redis nodes are being decreased
	ConditionScalingRedisDown ConditionType = "ScalingRedisDown"
	// ConditionScalingSentinelUp defines that the sentinel nodes are being increased
	ConditionScalingSentinelUp ConditionType = "ScalingSentinelUp"
	// ConditionScalingSentinelDown defines that the sentinel nodes are being decreased
	ConditionScalingSentinelDown ConditionType = "ScalingSentinelDown"
)

// Condition saves the state information of the redisfailover
type Condition struct {
	Type           ConditionType `json:"type"`
	Reason         string        `json:"reason"`
	TransitionTime string        `json:"transitionTime"`
}

// RedisFailoverStatus has the status of the cluster
type RedisFailoverStatus struct {
	Phase      Phase       `json:"phase"`
	Conditions []Condition `json:"conditions"`
	Master     string      `json:"master"`
}

// SetPhase sets the actual phase of the RF
func (s *RedisFailoverStatus) SetPhase(p Phase) {
	s.Phase = p
}

// SetReadyCondition saves the ready condition into the RF status
func (s *RedisFailoverStatus) SetReadyCondition() {
	condition := Condition{
		Type:           ConditionReady,
		TransitionTime: now(),
	}

	if len(s.Conditions) == 0 {
		s.appendCondition(condition)
		return
	}

	lastc := s.Conditions[len(s.Conditions)-1]
	if lastc.Type == ConditionReady {
		return
	}
	s.appendCondition(condition)
}

// SetNotReadyCondition saves the NotReady condition into the RF status
func (s *RedisFailoverStatus) SetNotReadyCondition() {
	condition := Condition{
		Type:           ConditionNotReady,
		TransitionTime: now(),
	}

	if len(s.Conditions) == 0 {
		s.appendCondition(condition)
		return
	}

	lastc := s.Conditions[len(s.Conditions)-1]
	if lastc.Type == ConditionNotReady {
		return
	}
	s.appendCondition(condition)
}

// AppendUpdatingRedisCondition adds the UpdatingRedis condition into the RF status
func (s *RedisFailoverStatus) AppendUpdatingRedisCondition(reason string) {
	c := Condition{
		Type:           ConditionUpdatingRedis,
		Reason:         reason,
		TransitionTime: now(),
	}
	s.appendCondition(c)
}

// AppendUpdatingSentinelCondition adds the UpdatingSentinel condition into the RF status
func (s *RedisFailoverStatus) AppendUpdatingSentinelCondition(reason string) {
	c := Condition{
		Type:           ConditionUpdatingSentinel,
		Reason:         reason,
		TransitionTime: now(),
	}
	s.appendCondition(c)
}

// AppendScalingRedisUpCondition add the ScalingRedisUp condition into the RF status
func (s *RedisFailoverStatus) AppendScalingRedisUpCondition(from, to int32) {
	c := Condition{
		Type:           ConditionScalingRedisUp,
		Reason:         scalingReason(from, to),
		TransitionTime: now(),
	}
	s.appendCondition(c)
}

// AppendScalingRedisDownCondition adds the ScalingRedisDown condition into the RF status
func (s *RedisFailoverStatus) AppendScalingRedisDownCondition(from, to int32) {
	c := Condition{
		Type:           ConditionScalingRedisDown,
		Reason:         scalingReason(from, to),
		TransitionTime: now(),
	}
	s.appendCondition(c)
}

// AppendScalingSentinelUpCondition adds the ScalingSentinelUp condition into the RF status
func (s *RedisFailoverStatus) AppendScalingSentinelUpCondition(from, to int32) {
	c := Condition{
		Type:           ConditionScalingSentinelUp,
		Reason:         scalingReason(from, to),
		TransitionTime: now(),
	}
	s.appendCondition(c)
}

// AppendScalingSentinelDownCondition adds the ScalingSentinelDown condition into the RF status
func (s *RedisFailoverStatus) AppendScalingSentinelDownCondition(from, to int32) {
	c := Condition{
		Type:           ConditionScalingSentinelDown,
		Reason:         scalingReason(from, to),
		TransitionTime: now(),
	}
	s.appendCondition(c)
}

// SetMaster saves the actual master into the RF status
func (s *RedisFailoverStatus) SetMaster(m string) {
	s.Master = m
}

func (s *RedisFailoverStatus) appendCondition(condition Condition) {
	s.Conditions = append(s.Conditions, condition)
	if len(s.Conditions) > 10 {
		s.Conditions = s.Conditions[1:]
	}
}

func scalingReason(from, to int32) string {
	return fmt.Sprintf("Current size: %d, desired size: %d", from, to)
}

func now() string {
	return time.Now().Format(time.RFC3339)
}
