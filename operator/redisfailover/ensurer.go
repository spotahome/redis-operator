package redisfailover

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	redisfailoverv1alpha2 "github.com/spotahome/redis-operator/api/redisfailover/v1alpha2"
)

func (w *RedisFailoverHandler) Ensure(rf *redisfailoverv1alpha2.RedisFailover, labels map[string]string, or []metav1.OwnerReference) error {
	// logger := w.logger.WithField("name", rf.Name).WithField("namespace", rf.Namespace)
	if rf.Spec.Redis.Exporter {
		if err := w.rfService.EnsureRedisService(rf, labels, or); err != nil {
			return err
		}
	} else {
		if err := w.rfService.EnsureNotPresentRedisService(rf); err != nil {
			return err
		}
	}
	if err := w.rfService.EnsureSentinelService(rf, labels, or); err != nil {
		return err
	}
	if err := w.rfService.EnsureSentinelConfigMap(rf, labels, or); err != nil {
		return err
	}
	if err := w.rfService.EnsureRedisConfigMap(rf, labels, or); err != nil {
		return err
	}
	if err := w.rfService.EnsureRedisStatefulset(rf, labels, or); err != nil {
		return err
	}
	if err := w.rfService.EnsureSentinelDeployment(rf, labels, or); err != nil {
		return err
	}

	return nil
}
