package redisfailover

import (
	"errors"
	"time"

	redisfailoverv1alpha2 "github.com/spotahome/redis-operator/api/redisfailover/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeToPrepare = 2 * time.Minute
)

func (r *RedisFailoverHandler) CheckAndHeal(rf *redisfailoverv1alpha2.RedisFailover, rfs []metav1.OwnerReference) error {
	// Number of redis is equal as the set on the RF spec
	// Number of sentinel is equal as the set on the RF spec
	// Check only one master
	// Number of redis master is 1
	// All redis slaves have the same master
	// All sentinels points to the same redis master
	// Sentinel has not death nodes
	// Sentinel knows the correct slave number
	if err := r.rfChecker.CheckRedisNumber(rf); err != nil {
		r.logger.Debug("Number of redis mismatch, this could be for a change on the statefulset")
		return nil
	}
	if err := r.rfChecker.CheckSentinelNumber(rf); err != nil {
		r.logger.Debug("Number of sentinel mismatch, this could be for a change on the deployment")
		return nil
	}

	nMasters, err := r.rfChecker.GetNumberMasters(rf)
	if err != nil {
		return err
	}
	switch nMasters {
	case 0:
		redisesIP, err := r.rfChecker.GetRedisesIPs(rf)
		if err != nil {
			return err
		}
		if len(redisesIP) == 1 {
			if err := r.rfHealer.MakeMaster(redisesIP[0]); err != nil {
				return err
			}
			break
		}
		minTime, err2 := r.rfChecker.GetMinimumRedisPodTime(rf)
		if err2 != nil {
			return err2
		}
		if minTime > timeToPrepare {
			r.logger.Debugf("time %.f more than expected. Not even one master, fixing...", minTime.Round(time.Second).Seconds())
			// We can consider there's an error
			if err2 := r.rfHealer.SetOldestAsMaster(rf); err2 != nil {
				r.mClient.SetClusterError(rf.Namespace, rf.Name)
				return err2
			}
		} else {
			// We'll wait until failover is done
			r.logger.Debug("No master found, wait until failover")
			return nil
		}
	case 1:
		break
	default:
		r.mClient.SetClusterError(rf.Namespace, rf.Name)
		return errors.New("More than one master, fix manually")
	}

	master, err := r.rfChecker.GetMasterIP(rf)
	if err != nil {
		return err
	}
	if err2 := r.rfChecker.CheckAllSlavesFromMaster(master, rf); err2 != nil {
		r.logger.Debug("Not all slaves have the same master")
		if err3 := r.rfHealer.SetMasterOnAll(master, rf); err3 != nil {
			return err3
		}
	}
	if err := r.rfService.EnsureRedissService(rf, rfs); err != nil {
		return err
	}
	redises, err := r.rfChecker.GetRedisesIPs(rf)
	if err != nil {
		return err
	}
	for _, rip := range redises {
		if err := r.rfHealer.SetRedisCustomConfig(rip, rf); err != nil {
			return err
		}
	}

	sentinels, err := r.rfChecker.GetSentinelsIPs(rf)
	if err != nil {
		return err
	}
	for _, sip := range sentinels {
		if err := r.rfChecker.CheckSentinelMonitor(sip, master); err != nil {
			r.logger.Debug("Sentinel is not monitoring the correct master")
			if err := r.rfHealer.NewSentinelMonitor(sip, master, rf); err != nil {
				r.mClient.SetClusterError(rf.Namespace, rf.Name)
				return err
			}
		}
	}
	for _, sip := range sentinels {
		if err := r.rfChecker.CheckSentinelNumberInMemory(sip, rf); err != nil {
			r.logger.Debug("Sentinel has more sentinel in memory than spected")
			if err := r.rfHealer.RestoreSentinel(sip); err != nil {
				return err
			}
		}
	}
	for _, sip := range sentinels {
		if err := r.rfChecker.CheckSentinelSlavesNumberInMemory(sip, rf); err != nil {
			r.logger.Debug("Sentinel has more slaves in memory than spected")
			if err := r.rfHealer.RestoreSentinel(sip); err != nil {
				return err
			}
		}
	}
	for _, sip := range sentinels {
		if err := r.rfHealer.SetSentinelCustomConfig(sip, rf); err != nil {
			return err
		}
	}
	r.mClient.SetClusterOK(rf.Namespace, rf.Name)
	return nil
}
