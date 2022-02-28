package redisfailover

import (
	"errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/operator/redisfailover/util"
)

const (
	timeToPrepare = 2 * time.Minute
)

// AllRestarted returns whether all pods' creationTimestamps are after a time
func (r *RedisFailoverHandler) AllRestarted(rf *redisfailoverv1.RedisFailover, pods []string, restartAt *metav1.Time) bool {
	if restartAt == nil {
		return true
	}
	for _, p := range pods {
		if ct, err := r.rfChecker.GetPodCreationTimestamp(p, rf); err != nil || restartAt.After(ct.Time) {
			return false
		}
	}
	return true
}

//UpdateRedisesPods if the running version of pods are equal to the statefulset one
func (r *RedisFailoverHandler) UpdateRedisesPods(rf *redisfailoverv1.RedisFailover) error {
	redises, err := r.rfChecker.GetRedisesIPs(rf)
	if err != nil {
		return err
	}

	redisNeedsRestart := util.RedisNeedsRestart(rf)

	masterIP := ""
	if !rf.Bootstrapping() {
		masterIP, _ = r.rfChecker.GetMasterIP(rf)
	}
	// No perform updates when nodes are syncing, still not connected, etc.
	for _, rp := range redises {
		if rp != masterIP {
			ready, err := r.rfChecker.CheckRedisSlavesReady(rp, rf)
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
	rfrRestartAt := rf.Spec.Redis.RestartAt
	for _, pod := range redisesPods {
		revision, err := r.rfChecker.GetRedisRevisionHash(pod, rf)
		if err != nil {
			return err
		}

		creationTimestamp, err := r.rfChecker.GetPodCreationTimestamp(pod, rf)
		if err != nil {
			return err
		}

		if revision != ssUR || (redisNeedsRestart && rfrRestartAt.After(creationTimestamp.Time)) {
			//Delete pod and wait next round to check if the new one is synced
			err = r.rfHealer.DeletePod(pod, rf)
			if err != nil {
				return err
			}
			if rf.Bootstrapping() {
				newrps, err := r.rfChecker.GetRedisesSlavesPods(rf)
				if err != nil {
					return err
				}
				if r.AllRestarted(rf, newrps, rfrRestartAt) {
					if rfrRestartAt == nil {
						t := metav1.Now()
						rfrRestartAt = &t
					}
					r.rfService.UpdateRedisRestartedAt(rf, &rfrRestartAt.Time)
				}
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

		masterCreationTimestamp, err := r.rfChecker.GetPodCreationTimestamp(master, rf)
		if err != nil {
			return err
		}

		rfrRestartAt := rf.Spec.Redis.RestartAt
		if masterRevision != ssUR || (redisNeedsRestart && rfrRestartAt.After(masterCreationTimestamp.Time)) {
			err = r.rfHealer.DeletePod(master, rf)
			if err != nil {
				return err
			}
			if rfrRestartAt == nil {
				t := metav1.Now()
				rfrRestartAt = &t
			}
			r.rfService.UpdateRedisRestartedAt(rf, &rfrRestartAt.Time)
			return nil
		}
	}

	return nil
}

// CheckAndRestartSentinels will check if sentinels a restart based on the pod creationTimestamp and RedisFailover's sentinel restartAt
func (r *RedisFailoverHandler) CheckAndRestartSentinels(rf *redisfailoverv1.RedisFailover) error {
	if !util.SentinelNeedsRestart(rf) {
		return nil
	}

	sps, err := r.rfChecker.GetSentinelsPods(rf)
	if err != nil {
		return err
	}

	rfsRestartAt := rf.Spec.Sentinel.RestartAt

	for _, sp := range sps {
		creationTimestamp, err := r.rfChecker.GetPodCreationTimestamp(sp, rf)
		if err != nil {
			return err
		}

		if creationTimestamp.After(rfsRestartAt.Time) || creationTimestamp.Equal(rfsRestartAt) {
			continue
		} else {
			if err := r.rfHealer.DeletePod(sp, rf); err != nil {
				return err
			}
			newsps, err := r.rfChecker.GetSentinelsPods(rf)
			if err != nil {
				return err
			}
			if r.AllRestarted(rf, newsps, rfsRestartAt) {
				if rfsRestartAt == nil {
					t := metav1.Now()
					rfsRestartAt = &t
				}
				r.rfService.UpdateSentinelRestartedAt(rf, &rfsRestartAt.Time)
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
		break
	default:
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

	if err := r.applyRedisCustomConfig(rf); err != nil {
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
	for _, sip := range sentinels {
		if err := r.rfChecker.CheckSentinelMonitor(sip, master); err != nil {
			r.logger.Debug("Sentinel is not monitoring the correct master")
			if err := r.rfHealer.NewSentinelMonitor(sip, master, rf); err != nil {
				return err
			}
		}
	}
	r.CheckAndRestartSentinels(rf)
	return r.checkAndHealSentinels(rf, sentinels)
}

func (r *RedisFailoverHandler) checkAndHealBootstrapMode(rf *redisfailoverv1.RedisFailover) error {
	if err := r.rfChecker.CheckRedisNumber(rf); err != nil {
		r.logger.Debug("Number of redis mismatch, this could be for a change on the statefulset")
		return nil
	}

	err := r.UpdateRedisesPods(rf)
	if err != nil {
		return err
	}

	if err := r.applyRedisCustomConfig(rf); err != nil {
		return err
	}

	bootstrapSettings := rf.Spec.BootstrapNode
	if err := r.rfHealer.SetExternalMasterOnAll(bootstrapSettings.Host, bootstrapSettings.Port, rf); err != nil {
		return err
	}

	if rf.SentinelsAllowed() {
		if err := r.rfChecker.CheckSentinelNumber(rf); err != nil {
			r.logger.Debug("Number of sentinel mismatch, this could be for a change on the deployment")
			return nil
		}

		sentinels, err := r.rfChecker.GetSentinelsIPs(rf)
		if err != nil {
			return err
		}
		for _, sip := range sentinels {
			if err := r.rfChecker.CheckSentinelMonitor(sip, bootstrapSettings.Host, bootstrapSettings.Port); err != nil {
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
	return nil
}
