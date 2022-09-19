package service

import (
	"errors"
	"sort"
	"strconv"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	redisauth "github.com/spotahome/redis-operator/operator/redisfailover/auth"
	"github.com/spotahome/redis-operator/service/k8s"
	"github.com/spotahome/redis-operator/service/redis"
	v1 "k8s.io/api/core/v1"
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

func (r *RedisFailoverHealer) setMasterLabelIfNecessary(namespace string, pod v1.Pod) error {
	for labelKey, labelValue := range pod.ObjectMeta.Labels {
		if labelKey == redisRoleLabelKey && labelValue == redisRoleLabelMaster {
			return nil
		}
	}
	return r.k8sService.UpdatePodLabels(namespace, pod.ObjectMeta.Name, generateRedisMasterRoleLabel())
}

func (r *RedisFailoverHealer) setSlaveLabelIfNecessary(namespace string, pod v1.Pod) error {
	for labelKey, labelValue := range pod.ObjectMeta.Labels {
		if labelKey == redisRoleLabelKey && labelValue == redisRoleLabelSlave {
			return nil
		}
	}
	return r.k8sService.UpdatePodLabels(namespace, pod.ObjectMeta.Name, generateRedisSlaveRoleLabel())
}

func (r *RedisFailoverHealer) MakeMaster(ip string, rf *redisfailoverv1.RedisFailover) error {

	authProvider := redisauth.GetAuthProvider(rf, r.k8sService)
	username, password, err := authProvider.GetAdminCredentials()
	if err != nil {
		return err
	}

	port := getRedisPort(rf.Spec.Redis.Port)
	err = r.redisClient.MakeMaster(ip, port, username, password)
	if err != nil {
		return err
	}

	rps, err := r.k8sService.GetStatefulSetPods(rf.Namespace, GetRedisName(rf))
	if err != nil {
		return err
	}
	for _, rp := range rps.Items {
		if rp.Status.PodIP == ip {
			return r.setMasterLabelIfNecessary(rf.Namespace, rp)
		}
	}
	return nil
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

	authProvider := redisauth.GetAuthProvider(rf, r.k8sService)
	username, password, err := authProvider.GetAdminCredentials()
	if err != nil {
		return err
	}

	port := getRedisPort(rf.Spec.Redis.Port)
	newMasterIP := ""
	for _, pod := range ssp.Items {
		if newMasterIP == "" {
			newMasterIP = pod.Status.PodIP
			r.logger.Debugf("New master is %s with ip %s", pod.Name, newMasterIP)
			if err := r.redisClient.MakeMaster(newMasterIP, port, username, password); err != nil {
				r.logger.Errorf("Make new master failed, master ip: %s, error: %v", pod.Status.PodIP, err)
				continue
			}

			err = r.setMasterLabelIfNecessary(rf.Namespace, pod)
			if err != nil {
				return err
			}

			newMasterIP = pod.Status.PodIP
		} else {
			r.logger.Debugf("Making pod %s slave of %s", pod.Name, newMasterIP)
			if err := r.redisClient.MakeSlaveOfWithPort(pod.Status.PodIP, newMasterIP, port, username, password); err != nil {
				r.logger.Errorf("Make slave failed, slave pod ip: %s, master ip: %s, error: %v", pod.Status.PodIP, newMasterIP, err)
			}

			err = r.setSlaveLabelIfNecessary(rf.Namespace, pod)
			if err != nil {
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

	authProvider := redisauth.GetAuthProvider(rf, r.k8sService)
	username, password, err := authProvider.GetAdminCredentials()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	port := getRedisPort(rf.Spec.Redis.Port)
	for _, pod := range ssp.Items {
		if pod.Status.PodIP == masterIP {
			r.logger.Debugf("Ensure pod %s is master", pod.Name)
			if err := r.redisClient.MakeMaster(masterIP, port, username, password); err != nil {
				r.logger.Errorf("Make master failed, master ip: %s, error: %v", masterIP, err)
				return err
			}

			err = r.setMasterLabelIfNecessary(rf.Namespace, pod)
			if err != nil {
				return err
			}
		} else {
			r.logger.Debugf("Making pod %s slave of %s", pod.Name, masterIP)
			if err := r.redisClient.MakeSlaveOfWithPort(pod.Status.PodIP, masterIP, port, username, password); err != nil {
				r.logger.Errorf("Make slave failed, slave ip: %s, master ip: %s, error: %v", pod.Status.PodIP, masterIP, err)
				return err
			}

			err = r.setSlaveLabelIfNecessary(rf.Namespace, pod)
			if err != nil {
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

	authProvider := redisauth.GetAuthProvider(rf, r.k8sService)
	username, password, err := authProvider.GetAdminCredentials()
	if err != nil {
		return err
	}

	for _, pod := range ssp.Items {
		r.logger.Debugf("Making pod %s slave of %s:%s", pod.Name, masterIP, masterPort)
		if err := r.redisClient.MakeSlaveOfWithPort(pod.Status.PodIP, masterIP, masterPort, username, password); err != nil {
			return err
		}

	}
	return nil
}

// NewSentinelMonitor changes the master that Sentinel has to monitor
func (r *RedisFailoverHealer) NewSentinelMonitor(ip string, monitor string, rf *redisfailoverv1.RedisFailover) error {
	r.logger.Debug("Sentinel is not monitoring the correct master, changing...")
	quorum := strconv.Itoa(int(getQuorum(rf)))

	authProvider := redisauth.GetAuthProvider(rf, r.k8sService)
	username, password, err := authProvider.GetAdminCredentials()
	if err != nil {
		return err
	}

	port := getRedisPort(rf.Spec.Redis.Port)
	return r.redisClient.MonitorRedisWithPort(ip, monitor, port, quorum, username, password)
}

// NewSentinelMonitorWithPort changes the master that Sentinel has to monitor by the provided IP and Port
func (r *RedisFailoverHealer) NewSentinelMonitorWithPort(ip string, monitor string, monitorPort string, rf *redisfailoverv1.RedisFailover) error {
	r.logger.Debug("Sentinel is not monitoring the correct master, changing...")
	quorum := strconv.Itoa(int(getQuorum(rf)))

	authProvider := redisauth.GetAuthProvider(rf, r.k8sService)
	username, password, err := authProvider.GetAdminCredentials()
	if err != nil {
		return err
	}

	return r.redisClient.MonitorRedisWithPort(ip, monitor, monitorPort, quorum, username, password)
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

	authProvider := redisauth.GetAuthProvider(rf, r.k8sService)
	username, password, err := authProvider.GetAdminCredentials()
	if err != nil {
		return err
	}

	port := getRedisPort(rf.Spec.Redis.Port)
	return r.redisClient.SetCustomRedisConfig(ip, port, rf.Spec.Redis.CustomConfig, username, password)
}

// DeletePod delete a failing pod so kubernetes relaunch it again
func (r *RedisFailoverHealer) DeletePod(podName string, rFailover *redisfailoverv1.RedisFailover) error {
	r.logger.Debugf("Deleting pods %s...", podName)
	return r.k8sService.DeletePod(rFailover.Namespace, podName)
}
