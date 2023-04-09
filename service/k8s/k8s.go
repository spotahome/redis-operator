package k8s

import (
	apiextensionscli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	redisfailoverclientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
)

// Service is the K8s service entrypoint.
type Services interface {
	ConfigMap
	Secret
	Pod
	PodDisruptionBudget
	RedisFailover
	Service
	RBAC
	Deployment
	StatefulSet
}

type services struct {
	ConfigMap
	Secret
	Pod
	PodDisruptionBudget
	RedisFailover
	Service
	RBAC
	Deployment
	StatefulSet
}

// New returns a new Kubernetes service.
func New(kubecli kubernetes.Interface, crdcli redisfailoverclientset.Interface, apiextcli apiextensionscli.Interface, logger log.Logger, metricsRecorder metrics.Recorder, useCache bool) Services {
	return &services{
		ConfigMap:           NewConfigMapService(kubecli, logger, metricsRecorder, useCache),
		Secret:              NewSecretService(kubecli, logger, metricsRecorder, useCache),
		Pod:                 NewPodService(kubecli, logger, metricsRecorder, useCache),
		PodDisruptionBudget: NewPodDisruptionBudgetService(kubecli, logger, metricsRecorder, useCache),
		RedisFailover:       NewRedisFailoverService(crdcli, logger, metricsRecorder),
		Service:             NewServiceService(kubecli, logger, metricsRecorder, useCache),
		RBAC:                NewRBACService(kubecli, logger, metricsRecorder, useCache),
		Deployment:          NewDeploymentService(kubecli, logger, metricsRecorder, useCache),
		StatefulSet:         NewStatefulSetService(kubecli, logger, metricsRecorder, useCache),
	}
}
