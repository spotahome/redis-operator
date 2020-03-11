package service

import (
	"errors"
	"sort"
	"strconv"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/service/k8s"
	"github.com/spotahome/redis-operator/service/redis"
)

// RedisFailoverHeal defines the interface able to fix the problems on the redis failovers
type RedisFailoverHeal interface {
	MakeMaster(ip string, rFailover *redisfailoverv1.RedisFailover) error
	SetOldestAsMaster(rFailover *redisfailoverv1.RedisFailover) error
	SetMasterOnAll(masterIP string, rFailover *redisfailoverv1.RedisFailover) error
	SetExternalMasterOnAll(masterIP string, masterPort string, rFailover *redisfailoverv1.RedisFailover) error
	NewSentinelMonitor(ip string, monitor string, rFailover *redisfailoverv1.RedisFailover) error
	NewSentinelMonitorWithPort(ip string, monitor string, port string, rFailover *redisfailoverv1.RedisFailover) error
	RestoreSentinel(ip string) error
	SetSentinelCustomConfig(ip string, rFailover *redisfailoverv1.RedisFailover) error
	SetRedisCustomConfig(ip string, rFailover *redisfailoverv1.RedisFailover) error
	DeletePod(podName string, rFailover *redisfailoverv1.RedisFailover) error
}

// RedisFailoverHealer is our implementation of RedisFailoverCheck interface
type RedisFailoverHealer struct {
	k8sService  k8s.Services
	redisClient redis.Client
	logger      log.Logger
}

// NewRedisFailoverHealer creates an object of the RedisFailoverChecker struct
func NewRedisFailoverHealer(k8sService k8s.Services, redisClient redis.Client, logger log.Logger) *RedisFailoverHealer {
	return &RedisFailoverHealer{
		k8sService:  k8sService,
		redisClient: redisClient,
		logger:      logger,
	}
}

func (r *RedisFailoverHealer) MakeMaster(ip string, rf *redisfailoverv1.RedisFailover) error {
	password, err := k8s.GetRedisPassword(r.k8sService, rf)
	if err != nil {
		return err
	}

	return r.redisClient.MakeMaster(ip, password)
}

// SetOldestAsMaster puts all redis to the same master, choosen by order of appearance
func (r *RedisFailoverHealer) SetOldestAsMaster(rf *redisfailoverv1.RedisFailover) error {
	ssp, err := r.k8sService.GetStatefulSetPods(rf.Namespace, GetRedisName(rf))
	if err != nil {
		return err
	}
	if len(ssp.Items) < 1 {
		return errors.New("number of redis pods are 0")
	}

	// Order the pods so we start by the oldest one
	sort.Slice(ssp.Items, func(i, j int) bool {
		return ssp.Items[i].CreationTimestamp.Before(&ssp.Items[j].CreationTimestamp)
	})

	password, err := k8s.GetRedisPassword(r.k8sService, rf)
	if err != nil {
		return err
	}

	newMasterIP := ""
	for _, pod := range ssp.Items {
		if newMasterIP == "" {
			newMasterIP = pod.Status.PodIP
			r.logger.Debugf("New master is %s with ip %s", pod.Name, newMasterIP)
			if err := r.redisClient.MakeMaster(newMasterIP, password); err != nil {
				return err
			}
		} else {
			r.logger.Debugf("Making pod %s slave of %s", pod.Name, newMasterIP)
			if err := r.redisClient.MakeSlaveOf(pod.Status.PodIP, newMasterIP, password); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetMasterOnAll puts all redis nodes as a slave of a given master
func (r *RedisFailoverHealer) SetMasterOnAll(masterIP string, rf *redisfailoverv1.RedisFailover) error {
	ssp, err := r.k8sService.GetStatefulSetPods(rf.Namespace, GetRedisName(rf))
	if err != nil {
		return err
	}

	password, err := k8s.GetRedisPassword(r.k8sService, rf)
	if err != nil {
		return err
	}

	for _, pod := range ssp.Items {
		if pod.Status.PodIP == masterIP {
			r.logger.Debugf("Ensure pod %s is master", pod.Name)
			if err := r.redisClient.MakeMaster(masterIP, password); err != nil {
				return err
			}
		} else {
			r.logger.Debugf("Making pod %s slave of %s", pod.Name, masterIP)
			if err := r.redisClient.MakeSlaveOf(pod.Status.PodIP, masterIP, password); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetExternalMasterOnAll puts all redis nodes as a slave of a given master outside of
// the current RedisFailover instance
func (r *RedisFailoverHealer) SetExternalMasterOnAll(masterIP, masterPort string, rf *redisfailoverv1.RedisFailover) error {
	ssp, err := r.k8sService.GetStatefulSetPods(rf.Namespace, GetRedisName(rf))
	if err != nil {
		return err
	}

	password, err := k8s.GetRedisPassword(r.k8sService, rf)
	if err != nil {
		return err
	}

	for _, pod := range ssp.Items {
		r.logger.Debugf("Making pod %s slave of %s:%s", pod.Name, masterIP, masterPort)
		if err := r.redisClient.MakeSlaveOfWithPort(pod.Status.PodIP, masterIP, masterPort, password); err != nil {
			return err
		}

	}
	return nil
}

// NewSentinelMonitor changes the master that Sentinel has to monitor
func (r *RedisFailoverHealer) NewSentinelMonitor(ip string, monitor string, rf *redisfailoverv1.RedisFailover) error {
	r.logger.Debug("Sentinel is not monitoring the correct master, changing...")
	quorum := strconv.Itoa(int(getQuorum(rf)))

	password, err := k8s.GetRedisPassword(r.k8sService, rf)
	if err != nil {
		return err
	}

	return r.redisClient.MonitorRedis(ip, monitor, quorum, password)
}

// NewSentinelMonitorWithPort changes the master that Sentinel has to monitor by the provided IP and Port
func (r *RedisFailoverHealer) NewSentinelMonitorWithPort(ip string, monitor string, monitorPort string, rf *redisfailoverv1.RedisFailover) error {
	r.logger.Debug("Sentinel is not monitoring the correct master, changing...")
	quorum := strconv.Itoa(int(getQuorum(rf)))

	password, err := k8s.GetRedisPassword(r.k8sService, rf)
	if err != nil {
		return err
	}

	return r.redisClient.MonitorRedisWithPort(ip, monitor, monitorPort, quorum, password)
}

// RestoreSentinel clear the number of sentinels on memory
func (r *RedisFailoverHealer) RestoreSentinel(ip string) error {
	r.logger.Debugf("Restoring sentinel %s...", ip)
	return r.redisClient.ResetSentinel(ip)
}

// SetSentinelCustomConfig will call sentinel to set the configuration given in config
func (r *RedisFailoverHealer) SetSentinelCustomConfig(ip string, rf *redisfailoverv1.RedisFailover) error {
	r.logger.Debugf("Setting the custom config on sentinel %s...", ip)
	return r.redisClient.SetCustomSentinelConfig(ip, rf.Spec.Sentinel.CustomConfig)
}

// SetRedisCustomConfig will call redis to set the configuration given in config
func (r *RedisFailoverHealer) SetRedisCustomConfig(ip string, rf *redisfailoverv1.RedisFailover) error {
	r.logger.Debugf("Setting the custom config on redis %s...", ip)

	password, err := k8s.GetRedisPassword(r.k8sService, rf)
	if err != nil {
		return err
	}

	return r.redisClient.SetCustomRedisConfig(ip, rf.Spec.Redis.CustomConfig, password)
}

//DeletePod delete a failing pod so kubernetes relaunch it again
func (r *RedisFailoverHealer) DeletePod(podName string, rFailover *redisfailoverv1.RedisFailover) error {
	r.logger.Debugf("Deleting pods %s...", podName)
	return r.k8sService.DeletePod(rFailover.Namespace, podName)
}
