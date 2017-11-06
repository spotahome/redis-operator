package failover

import (
	"errors"
	"fmt"
	"time"

	"github.com/spotahome/redis-operator/pkg/clock"
	"github.com/spotahome/redis-operator/pkg/log"
	"github.com/spotahome/redis-operator/pkg/metrics"
	"github.com/spotahome/redis-operator/pkg/redis"
)

// RedisFailoverCheck defines the interface able to check the correct status of a redis failover
type RedisFailoverCheck interface {
	Check(rFailover *RedisFailover) error
	GetMaster(rFailover *RedisFailover) (string, error)
}

// RedisFailoverChecker is our implementation of RedisFailoverCheck interface
type RedisFailoverChecker struct {
	client      RedisFailoverClient
	redisClient redis.Client
	clock       clock.Clock
	metrics     metrics.Instrumenter
	logger      log.Logger
}

// NewRedisFailoverChecker creates an object of the RedisFailoverChecker struct
func NewRedisFailoverChecker(metricsClient metrics.Instrumenter, client RedisFailoverClient, redisClient redis.Client, clock clock.Clock, logger log.Logger) *RedisFailoverChecker {
	return &RedisFailoverChecker{
		client:      client,
		redisClient: redisClient,
		clock:       clock,
		metrics:     metricsClient,
		logger:      logger,
	}
}

// Check connects to the redis and sentinel nodes and checks if the following requirements are ok:
// * Number of redis is equal as the set on the RF spec
// * Number of sentinel is equal as the set on the RF spec
// * Number of redis master is 1
// * All redis slaves have the same master
// * All sentinels points to the same redis master
// * Sentinel has not death nodes
func (r *RedisFailoverChecker) Check(rf *RedisFailover) error {
	rNumber, err := r.checkRedisNumber(rf)
	if err != nil {
		return err
	} else if !rNumber {
		return errors.New("Redis number mismatch spec")
	}
	sNumber, err := r.checkSentinelNumber(rf)
	if err != nil {
		return err
	} else if !sNumber {
		return errors.New("Sentinel number mismatch spec")
	}
	master, err := r.GetMaster(rf)
	if err != nil {
		return err
	}
	if _, err := r.checkMasterAndSlaves(master, rf); err != nil {
		return err
	}
	if _, err := r.checkSentinelNumberInMemory(rf); err != nil {
		return err
	}
	return nil
}

// checkRedisNumber controlls that the number of deployed redis is the same than the requested on the spec
func (r *RedisFailoverChecker) checkRedisNumber(rf *RedisFailover) (bool, error) {
	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)
	ss, err := r.client.GetRedisStatefulset(rf)
	if err != nil {
		return false, err
	}
	logger.Debugf("Redis spected replicas: %d, StatefulSet replicas: %d", rf.Spec.Redis.Replicas, *ss.Spec.Replicas)
	return rf.Spec.Redis.Replicas == *ss.Spec.Replicas, nil
}

// checkSentinelNumber controlls that the number of deployed sentinel is the same than the requested on the spec
func (r *RedisFailoverChecker) checkSentinelNumber(rf *RedisFailover) (bool, error) {
	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)
	d, err := r.client.GetSentinelDeployment(rf)
	if err != nil {
		return false, err
	}
	logger.Debugf("Sentinel spected replicas: %d, Deployment replicas: %d", rf.Spec.Sentinel.Replicas, *d.Spec.Replicas)
	return rf.Spec.Sentinel.Replicas == *d.Spec.Replicas, nil
}

// checkMasterAndSlaves controlls that all slaves have the same master (the real one)
func (r *RedisFailoverChecker) checkMasterAndSlaves(master string, rf *RedisFailover) (bool, error) {
	rps, err := r.client.GetRedisPodsIPs(rf)
	if err != nil {
		return false, err
	}
	for _, redisNode := range rps {
		slave, err := r.redisClient.GetSlaveOf(redisNode)
		if err != nil {
			return false, err
		}
		if slave != "" && slave != master {
			return false, fmt.Errorf("Slave %s don't have the master %s, has %s", redisNode, master, slave)
		}
	}
	return true, nil
}

// checkSentinelNumberInMemory controls that sentinels have only the living sentinels on its memory.
// If don't, it restart the memory of that sentinel
func (r *RedisFailoverChecker) checkSentinelNumberInMemory(rf *RedisFailover) (bool, error) {
	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)
	sps, err := r.client.GetSentinelPodsIPs(rf)
	if err != nil {
		return false, err
	}
	for _, sp := range sps {
		nSentinels, err := r.redisClient.GetNumberSentinelsInMemory(sp)
		if err == nil && nSentinels != rf.Spec.Sentinel.Replicas {
			logger.Debugf("Sentinel has %d nodes, reseting...", nSentinels)
			err := r.redisClient.ResetSentinel(sp)
			if err != nil {
				return false, err
			}
			r.clock.Sleep(30 * time.Second)
		}
	}
	return true, nil
}

// GetMaster connects to all redis and returns the master of the redis failover
func (r *RedisFailoverChecker) GetMaster(rf *RedisFailover) (string, error) {
	rps, err := r.client.GetRedisPodsIPs(rf)
	if err != nil {
		return "", err
	}
	masters := []string{}
	for _, redisNode := range rps {
		master, err := r.redisClient.IsMaster(redisNode)
		if err != nil {
			return "", err
		}
		if master {
			masters = append(masters, redisNode)
		}
	}

	mLen := len(masters)
	r.metrics.SetClusterMasters(float64(mLen), rf.Metadata.Name)
	if mLen != 1 {
		return "", errors.New("Number of redis nodes known as master is different than 1")
	}
	return masters[0], nil
}
