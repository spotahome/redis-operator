package redisfailover

import (
	"errors"
	"strconv"
	"time"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/metrics"
)

const (
	timeToPrepare = 2 * time.Minute
)

// UpdateRedisesPods if the running version of pods are equal to the statefulset one
func (r *RedisFailoverHandler) UpdateRedisesPods(rf *redisfailoverv1.RedisFailover) error {
	redises, err := r.rfChecker.GetRedisesIPs(rf)
	if err != nil {
		return err
	}

	masterIP := ""
	if !rf.Bootstrapping() {
		masterIP, _ = r.rfChecker.GetMasterIP(rf)
	}
	// No perform updates when nodes are syncing, still not connected, etc.
	for _, rip := range redises {
		if rip != masterIP {
			ready, err := r.rfChecker.CheckRedisSlavesReady(rip, rf)
			if err != nil {
				return err
			}
			if !ready {
				return nil
			}
		}
	}

	ssUR, err := r.rfChecker.GetStatefulSetUpdateRevision(rf)
	if err != nil {
		return err
	}

	redisesPods, err := r.rfChecker.GetRedisesSlavesPods(rf)
	if err != nil {
		return err
	}

	// Update stale pods with slave role
	for _, pod := range redisesPods {
		revision, err := r.rfChecker.GetRedisRevisionHash(pod, rf)
		if err != nil {
			return err
		}
		if revision != ssUR {
			//Delete pod and wait next round to check if the new one is synced
			err = r.rfHealer.DeletePod(pod, rf)
			if err != nil {
				return err
			}
			return nil
		}
	}

	if !rf.Bootstrapping() {
		// Update stale pod with role master
		master, err := r.rfChecker.GetRedisesMasterPod(rf)
		if err != nil {
			return err
		}

		masterRevision, err := r.rfChecker.GetRedisRevisionHash(master, rf)
		if err != nil {
			return err
		}
		if masterRevision != ssUR {
			err = r.rfHealer.DeletePod(master, rf)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

// CheckAndHeal runs verifcation checks to ensure the RedisFailover is in an expected and healthy state.
// If the checks do not match up to expectations, an attempt will be made to "heal" the RedisFailover into a healthy state.
func (r *RedisFailoverHandler) CheckAndHeal(rf *redisfailoverv1.RedisFailover) error {
	if rf.Bootstrapping() {
		return r.checkAndHealBootstrapMode(rf)
	}

	// Number of redis is equal as the set on the RF spec
	// Number of sentinel is equal as the set on the RF spec
	// Check only one master
	// Number of redis master is 1
	// All redis slaves have the same master
	// All sentinels points to the same redis master
	// Sentinel has not death nodes
	// Sentinel knows the correct slave number

	err := r.rfChecker.CheckRedisNumber(rf)
	setRedisCheckerMetrics(r.mClient, "redis", rf.Namespace, rf.Name, metrics.REDIS_REPLICA_MISMATCH, metrics.NOT_APPLICABLE, err)
	if err != nil {
		r.logger.Debug("Number of redis mismatch, this could be for a change on the statefulset")
		return nil
	}
	r.mClient.RecordRedisCheck(rf.Namespace, rf.Name, metrics.REDIS_REPLICA_MISMATCH, metrics.NOT_APPLICABLE, metrics.STATUS_HEALTHY)

	err = r.rfChecker.CheckSentinelNumber(rf)
	setRedisCheckerMetrics(r.mClient, "sentinel", rf.Namespace, rf.Name, metrics.SENTINEL_REPLICA_MISMATCH, metrics.NOT_APPLICABLE, err)
	if err != nil {
		r.logger.Debug("Number of sentinel mismatch, this could be for a change on the deployment")
		return nil
	}

	nMasters, err := r.rfChecker.GetNumberMasters(rf)
	if err != nil {
		return err
	}
	switch nMasters {
	case 0:
		setRedisCheckerMetrics(r.mClient, "redis", rf.Namespace, rf.Name, metrics.NUMBER_OF_MASTERS, metrics.NOT_APPLICABLE, errors.New("No masters detected"))
		redisesIP, err := r.rfChecker.GetRedisesIPs(rf)
		if err != nil {
			return err
		}
		if len(redisesIP) == 1 {
			if err := r.rfHealer.MakeMaster(redisesIP[0], rf); err != nil {
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
				return err2
			}
		} else {
			// We'll wait until failover is done
			r.logger.Debug("No master found, wait until failover")
			return nil
		}
	case 1:
		setRedisCheckerMetrics(r.mClient, "redis", rf.Namespace, rf.Name, metrics.NUMBER_OF_MASTERS, metrics.NOT_APPLICABLE, nil)
	default:
		setRedisCheckerMetrics(r.mClient, "redis", rf.Namespace, rf.Name, metrics.NUMBER_OF_MASTERS, metrics.NOT_APPLICABLE, errors.New("Multiple masters detected"))
		return errors.New("More than one master, fix manually")
	}

	master, err := r.rfChecker.GetMasterIP(rf)
	if err != nil {
		return err
	}

	err2 := r.rfChecker.CheckAllSlavesFromMaster(master, rf)
	setRedisCheckerMetrics(r.mClient, "redis", rf.Namespace, rf.Name, metrics.SLAVE_WRONG_MASTER, metrics.NOT_APPLICABLE, err)
	if err2 != nil {
		r.logger.Debug("Not all slaves have the same master")
		if err3 := r.rfHealer.SetMasterOnAll(master, rf); err3 != nil {
			return err3
		}
	}

	err = r.applyRedisCustomConfig(rf)
	setRedisCheckerMetrics(r.mClient, "redis", rf.Namespace, rf.Name, metrics.APPLY_REDIS_CONFIG, metrics.NOT_APPLICABLE, err)
	if err != nil {
		return err
	}

	err = r.UpdateRedisesPods(rf)
	if err != nil {
		return err
	}

	sentinels, err := r.rfChecker.GetSentinelsIPs(rf)
	if err != nil {
		return err
	}

	port := getRedisPort(rf.Spec.Redis.Port)
	for _, sip := range sentinels {
		err = r.rfChecker.CheckSentinelMonitor(sip, master, port)
		setRedisCheckerMetrics(r.mClient, "sentinel", rf.Namespace, rf.Name, sip, metrics.NOT_APPLICABLE, err)
		if err != nil {
			r.logger.Debug("Sentinel is not monitoring the correct master")
			if err := r.rfHealer.NewSentinelMonitor(sip, master, rf); err != nil {
				return err
			}
		}

	}
	return r.checkAndHealSentinels(rf, sentinels)
}

func (r *RedisFailoverHandler) checkAndHealBootstrapMode(rf *redisfailoverv1.RedisFailover) error {
	err := r.rfChecker.CheckRedisNumber(rf)
	setRedisCheckerMetrics(r.mClient, "redis", rf.Namespace, rf.Name, metrics.REDIS_REPLICA_MISMATCH, metrics.NOT_APPLICABLE, err)
	if err != nil {
		r.logger.Debug("Number of redis mismatch, this could be for a change on the statefulset")
		return nil
	}

	err = r.UpdateRedisesPods(rf)
	if err != nil {
		return err
	}
	err = r.applyRedisCustomConfig(rf)
	setRedisCheckerMetrics(r.mClient, "redis", rf.Namespace, rf.Name, metrics.APPLY_REDIS_CONFIG, metrics.NOT_APPLICABLE, err)
	if err != nil {
		return err
	}

	bootstrapSettings := rf.Spec.BootstrapNode
	if err := r.rfHealer.SetExternalMasterOnAll(bootstrapSettings.Host, bootstrapSettings.Port, rf); err != nil {
		return err
	}

	if rf.SentinelsAllowed() {
		err = r.rfChecker.CheckSentinelNumber(rf)
		setRedisCheckerMetrics(r.mClient, "sentinel", rf.Namespace, rf.Name, metrics.SENTINEL_REPLICA_MISMATCH, metrics.NOT_APPLICABLE, err)
		if err != nil {
			return nil
		}

		sentinels, err := r.rfChecker.GetSentinelsIPs(rf)
		if err != nil {
			return err
		}
		for _, sip := range sentinels {
			err = r.rfChecker.CheckSentinelMonitor(sip, bootstrapSettings.Host, bootstrapSettings.Port)
			setRedisCheckerMetrics(r.mClient, "sentinel", rf.Namespace, rf.Name, metrics.SENTINEL_WRONG_MASTER, sip, err)
			if err != nil {
				r.logger.Debug("Sentinel is not monitoring the correct master")
				if err := r.rfHealer.NewSentinelMonitorWithPort(sip, bootstrapSettings.Host, bootstrapSettings.Port, rf); err != nil {
					return err
				}
			}
		}
		return r.checkAndHealSentinels(rf, sentinels)
	}
	return nil
}

func (r *RedisFailoverHandler) applyRedisCustomConfig(rf *redisfailoverv1.RedisFailover) error {
	redises, err := r.rfChecker.GetRedisesIPs(rf)
	if err != nil {
		return err
	}
	for _, rip := range redises {
		if err := r.rfHealer.SetRedisCustomConfig(rip, rf); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisFailoverHandler) checkAndHealSentinels(rf *redisfailoverv1.RedisFailover, sentinels []string) error {
	for _, sip := range sentinels {
		err := r.rfChecker.CheckSentinelNumberInMemory(sip, rf)
		setRedisCheckerMetrics(r.mClient, "sentinel", rf.Namespace, rf.Name, metrics.SENTINEL_NUMBER_IN_MEMORY_MISMATCH, sip, err)
		if err != nil {
			r.logger.Debug("Sentinel has more sentinel in memory than spected")
			if err := r.rfHealer.RestoreSentinel(sip); err != nil {
				return err
			}
		}

	}
	for _, sip := range sentinels {
		err := r.rfChecker.CheckSentinelSlavesNumberInMemory(sip, rf)
		setRedisCheckerMetrics(r.mClient, "sentinel", rf.Namespace, rf.Name, metrics.REDIS_SLAVES_NUMBER_IN_MEMORY_MISMATCH, sip, err)
		if err != nil {
			r.logger.Debug("Sentinel has more slaves in memory than spected")
			if err := r.rfHealer.RestoreSentinel(sip); err != nil {
				return err
			}
		}
	}
	for _, sip := range sentinels {
		err := r.rfHealer.SetSentinelCustomConfig(sip, rf)
		setRedisCheckerMetrics(r.mClient, "sentinel", rf.Namespace, rf.Name, metrics.APPLY_SENTINEL_CONFIG, sip, err)
		if err != nil {
			return err
		}
	}
	return nil
}

func getRedisPort(p int32) string {
	return strconv.Itoa(int(p))
}

func setRedisCheckerMetrics(metricsClient metrics.Recorder, mode /* redis or sentinel? */ string, rfNamespace string, rfName string, property string, IP string, err error) {
	if mode == "sentinel" {
		if err != nil {
			metricsClient.RecordSentinelCheck(rfNamespace, rfName, property, IP, metrics.STATUS_UNHEALTHY)
		} else {
			metricsClient.RecordSentinelCheck(rfNamespace, rfName, property, IP, metrics.STATUS_HEALTHY)
		}

	} else if mode == "redis" {
		if err != nil {
			metricsClient.RecordRedisCheck(rfNamespace, rfName, property, IP, metrics.STATUS_UNHEALTHY)
		} else {
			metricsClient.RecordRedisCheck(rfNamespace, rfName, property, IP, metrics.STATUS_HEALTHY)
		}
	}
}
