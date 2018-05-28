package service

import (
	"strconv"

	redisfailoverv1alpha2 "github.com/spotahome/redis-operator/api/redisfailover/v1alpha2"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/service/k8s"
	"github.com/spotahome/redis-operator/service/redis"
	"github.com/spotahome/redis-operator/operator/redisfailover/util"
)

// RedisFailoverHeal defines the interface able to fix the problems on the redis failovers
type RedisFailoverHeal interface {
	SetRandomMaster(rFailover *redisfailoverv1alpha2.RedisFailover) error
	SetMasterOnAll(masterIP string, rFailover *redisfailoverv1alpha2.RedisFailover) error
	NewSentinelMonitor(ip string, monitor string, rFailover *redisfailoverv1alpha2.RedisFailover) error
	RestoreSentinel(ip string) error
	SetRoleLable(masterIP string, rFailover *redisfailoverv1alpha2.RedisFailover) error
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

// SetRandomMaster puts all redis to the same master, choosen by order
func (r *RedisFailoverHealer) SetRandomMaster(rf *redisfailoverv1alpha2.RedisFailover) error {
	ssp, err := r.k8sService.GetStatefulSetPods(rf.Namespace, GetRedisName(rf))
	if err != nil {
		return err
	}
	newMasterIP := ""
	for _, pod := range ssp.Items {
		if newMasterIP == "" {
			newMasterIP = pod.Status.PodIP
			r.logger.Debugf("New master is %s with ip %s", pod.Name, newMasterIP)
			if err := r.redisClient.MakeMaster(newMasterIP); err != nil {
				return err
			}
		} else {
			r.logger.Debugf("Making pod %s slave of %s", pod.Name, newMasterIP)
			if err := r.redisClient.MakeSlaveOf(pod.Status.PodIP, newMasterIP); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetMasterOnAll puts all redis nodes as a slave of a given master
func (r *RedisFailoverHealer) SetMasterOnAll(masterIP string, rf *redisfailoverv1alpha2.RedisFailover) error {
	ssp, err := r.k8sService.GetStatefulSetPods(rf.Namespace, GetRedisName(rf))
	if err != nil {
		return err
	}
	for _, pod := range ssp.Items {
		if pod.Status.PodIP == masterIP {
			r.logger.Debugf("Ensure pod %s is master", pod.Name)
			if err := r.redisClient.MakeMaster(masterIP); err != nil {
				return err
			}
		} else {
			r.logger.Debugf("Making pod %s slave of %s", pod.Name, masterIP)
			if err := r.redisClient.MakeSlaveOf(pod.Status.PodIP, masterIP); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetRoleLable sets role tag for statful pod.
func (r *RedisFailoverHealer) SetRoleLable(masterIP string, rf *redisfailoverv1alpha2.RedisFailover) error {
	ssp, err := r.k8sService.GetStatefulSetPods(rf.Namespace, GetRedisName(rf))
	if err != nil {
		return err
	}
	redisRoleLable := map[string]string{redisRoleLabelName: ""}
	for _, pod := range ssp.Items{
		labels := pod.GetLabels()
		if pod.Status.PodIP == masterIP{
			r.logger.Debugf("Pod %s is master, updating pod labels as master", pod.Name)
			redisRoleLable[redisRoleLabelName] = redisMaster
		}else{
			r.logger.Debugf("Pod %s is slave, updating pod labels as salve", pod.Name)
			redisRoleLable[redisRoleLabelName] = redisSlave
		}
		labels = util.MergeLabels(labels, redisRoleLable)
		pod.SetLabels(labels)
	}
	return nil
}

// NewSentinelMonitor clear the number of sentinels on memory
func (r *RedisFailoverHealer) NewSentinelMonitor(ip string, monitor string, rf *redisfailoverv1alpha2.RedisFailover) error {
	r.logger.Debug("Sentinel is not monitoring the correct master, changing...")
	quorum := strconv.Itoa(int(getQuorum(rf)))
	return r.redisClient.MonitorRedis(ip, monitor, quorum)
}

// RestoreSentinel clear the number of sentinels on memory
func (r *RedisFailoverHealer) RestoreSentinel(ip string) error {
	r.logger.Debugf("Restoring sentinel %s...", ip)
	return r.redisClient.ResetSentinel(ip)
}
