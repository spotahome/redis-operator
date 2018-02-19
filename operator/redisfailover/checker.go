package redisfailover

import (
	"errors"
	"time"

	redisfailoverv1alpha2 "github.com/spotahome/redis-operator/api/redisfailover/v1alpha2"
)

const (
	timeToPrepare = 2 * time.Minute
)

func (w *RedisFailoverHandler) CheckAndHeal(rf *redisfailoverv1alpha2.RedisFailover) error {
	// Number of redis is equal as the set on the RF spec
	// Number of sentinel is equal as the set on the RF spec
	// Check only one master
	// Number of redis master is 1
	// All redis slaves have the same master
	// All sentinels points to the same redis master
	// Sentinel has not death nodes
	// Sentinel knows the correct slave number
	if err := w.rfChecker.CheckRedisNumber(rf); err != nil {
		w.logger.Debug("Number of redis mismatch, this could be for a change on the statefulset")
		return nil
	}
	if err := w.rfChecker.CheckSentinelNumber(rf); err != nil {
		w.logger.Debug("Number of sentinel mismatch, this could be for a change on the deployment")
		return nil
	}

	nMasters, err := w.rfChecker.GetNumberMasters(rf)
	if err != nil {
		return err
	}
	switch nMasters {
	case 0:
		minTime, err2 := w.rfChecker.GetMinimumRedisPodTime(rf)
		if err2 != nil {
			return err2
		}
		if minTime > timeToPrepare {
			w.logger.Debugf("Time %.f more than expected. Not even one master, fixing...", minTime.Round(time.Second).Seconds())
			// We can consider there's an error
			if err2 := w.rfHealer.SetRandomMaster(rf); err2 != nil {
				return err2
			}
		} else {
			// We'll wait until failover is done
			w.logger.Debug("No master found, wait until failover")
			return nil
		}
	case 1:
		break
	default:
		w.mClient.SetClusterError(rf.Namespace, rf.Name)
		return errors.New("More than one master, fix manually")
	}

	master, err := w.rfChecker.GetMasterIP(rf)
	if err != nil {
		return err
	}
	if err2 := w.rfChecker.CheckAllSlavesFromMaster(master, rf); err2 != nil {
		w.logger.Debug("Not all slaves have the same master")
		if err3 := w.rfHealer.SetMasterOnAll(master, rf); err3 != nil {
			return err3
		}
	}
	sentinels, err := w.rfChecker.GetSentinelsIPs(rf)
	if err != nil {
		return err
	}
	for _, sip := range sentinels {
		if err := w.rfChecker.CheckSentinelMonitor(sip, master); err != nil {
			w.logger.Debug("Sentinel is not monitoring the correct master")
			if err := w.rfHealer.NewSentinelMonitor(sip, master, rf); err != nil {
				return err
			}
		}
	}
	for _, sip := range sentinels {
		if err := w.rfChecker.CheckSentinelNumberInMemory(sip, rf); err != nil {
			w.logger.Debug("Sentinel has more sentinel in memory than spected")
			if err := w.rfHealer.RestoreSentinel(sip); err != nil {
				return err
			}
		}
	}
	for _, sip := range sentinels {
		if err := w.rfChecker.CheckSentinelSlavesNumberInMemory(sip, rf); err != nil {
			w.logger.Debug("Sentinel has more slaves in memory than spected")
			if err := w.rfHealer.RestoreSentinel(sip); err != nil {
				return err
			}
		}
	}
	w.mClient.SetClusterOK(rf.Namespace, rf.Name)
	return nil
}
