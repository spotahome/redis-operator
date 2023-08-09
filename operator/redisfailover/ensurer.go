package redisfailover

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/metrics"
)

// Ensure is called to ensure all of the resources associated with a RedisFailover are created
func (w *RedisFailoverHandler) Ensure(rf *redisfailoverv1.RedisFailover, labels map[string]string, or []metav1.OwnerReference, metricsClient metrics.Recorder) error {
	if rf.Spec.Redis.Exporter.Enabled {
		if err := w.rfService.EnsureRedisService(rf, labels, or); err != nil {
			return err
		}
		if err := w.rfService.EnsureRedisPodMonitor(rf, labels, or); err != nil {
			return err
		}
	} else {
		if err := w.rfService.EnsureNotPresentRedisService(rf); err != nil {
			return err
		}
		if err := w.rfService.EnsureNotPresentRedisPodMonitor(rf, labels, or); err != nil {
			return err
		}
	}

	if rf.Spec.Sentinel.Exporter.Enabled {
		if err := w.rfService.EnsureSentinelPodMonitor(rf, labels, or); err != nil {
			return err
		}
	} else {
		if err := w.rfService.EnsureNotPresentSentinelPodMonitor(rf, labels, or); err != nil {
			return err
		}
	}

	sentinelsAllowed := rf.SentinelsAllowed()
	if sentinelsAllowed {
		if err := w.rfService.EnsureSentinelService(rf, labels, or); err != nil {
			return err
		}
		if err := w.rfService.EnsureSentinelConfigMap(rf, labels, or); err != nil {
			return err
		}
	}

	if err := w.rfService.EnsureRedisShutdownConfigMap(rf, labels, or); err != nil {
		return err
	}
	if err := w.rfService.EnsureRedisReadinessConfigMap(rf, labels, or); err != nil {
		return err
	}
	if err := w.rfService.EnsureRedisConfigMap(rf, labels, or); err != nil {
		return err
	}
	if err := w.rfService.EnsureRedisStatefulset(rf, labels, or); err != nil {
		return err
	}

	if sentinelsAllowed {
		if err := w.rfService.EnsureSentinelDeployment(rf, labels, or); err != nil {
			return err
		}
	}

	return nil
}
