package service

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/operator/redisfailover/util"
	"github.com/spotahome/redis-operator/service/k8s"
)

// RedisFailoverClient has the minimumm methods that a Redis failover controller needs to satisfy
// in order to talk with K8s
type RedisFailoverClient interface {
	EnsureSentinelService(rFailover *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureSentinelConfigMap(rFailover *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureSentinelDeployment(rFailover *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisStatefulset(rFailover *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisService(rFailover *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisShutdownConfigMap(rFailover *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisReadinessConfigMap(rFailover *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureRedisConfigMap(rFailover *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error
	EnsureNotPresentRedisService(rFailover *redisfailoverv1.RedisFailover) error
}

// RedisFailoverKubeClient implements the required methods to talk with kubernetes
type RedisFailoverKubeClient struct {
	K8SService k8s.Services
	logger     log.Logger
}

// NewRedisFailoverKubeClient creates a new RedisFailoverKubeClient
func NewRedisFailoverKubeClient(k8sService k8s.Services, logger log.Logger) *RedisFailoverKubeClient {
	return &RedisFailoverKubeClient{
		K8SService: k8sService,
		logger:     logger,
	}
}

func generateSelectorLabels(component, name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      name,
		"app.kubernetes.io/component": component,
		"app.kubernetes.io/part-of":   appLabel,
	}
}

// EnsureSentinelService makes sure the sentinel service exists
func (r *RedisFailoverKubeClient) EnsureSentinelService(rf *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	svc := generateSentinelService(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsService(rf.Namespace, svc)
}

// EnsureSentinelConfigMap makes sure the sentinel configmap exists
func (r *RedisFailoverKubeClient) EnsureSentinelConfigMap(rf *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	cm := generateSentinelConfigMap(rf, labels, ownerRefs)
	return r.K8SService.CreateOrUpdateConfigMap(rf.Namespace, cm)
}

// EnsureSentinelDeployment makes sure the sentinel deployment exists in the desired state
func (r *RedisFailoverKubeClient) EnsureSentinelDeployment(rf *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	if err := r.ensurePodDisruptionBudget(rf, sentinelName, sentinelRoleName, labels, ownerRefs); err != nil {
		return err
	}
	d := generateSentinelDeployment(rf, labels, ownerRefs)
	return r.K8SService.CreateOrUpdateDeployment(rf.Namespace, d)
}

// EnsureRedisStatefulset makes sure the redis statefulset exists in the desired state
func (r *RedisFailoverKubeClient) EnsureRedisStatefulset(rf *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	if err := r.ensurePodDisruptionBudget(rf, redisName, redisRoleName, labels, ownerRefs); err != nil {
		return err
	}
	ss := generateRedisStatefulSet(rf, labels, ownerRefs)
	return r.K8SService.CreateOrUpdateStatefulSet(rf.Namespace, ss)
}

// EnsureRedisConfigMap makes sure the Redis ConfigMap exists
func (r *RedisFailoverKubeClient) EnsureRedisConfigMap(rf *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {

	password, err := k8s.GetRedisPassword(r.K8SService, rf)
	if err != nil {
		return err
	}

	cm := generateRedisConfigMap(rf, labels, ownerRefs, password)
	return r.K8SService.CreateOrUpdateConfigMap(rf.Namespace, cm)
}

// EnsureRedisShutdownConfigMap makes sure the redis configmap with shutdown script exists
func (r *RedisFailoverKubeClient) EnsureRedisShutdownConfigMap(rf *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	if rf.Spec.Redis.ShutdownConfigMap != "" {
		if _, err := r.K8SService.GetConfigMap(rf.Namespace, rf.Spec.Redis.ShutdownConfigMap); err != nil {
			return err
		}
	} else {
		cm := generateRedisShutdownConfigMap(rf, labels, ownerRefs)
		return r.K8SService.CreateOrUpdateConfigMap(rf.Namespace, cm)
	}
	return nil
}

// EnsureRedisReadinessConfigMap makes sure the redis configmap with shutdown script exists
func (r *RedisFailoverKubeClient) EnsureRedisReadinessConfigMap(rf *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	cm := generateRedisReadinessConfigMap(rf, labels, ownerRefs)
	return r.K8SService.CreateOrUpdateConfigMap(rf.Namespace, cm)
}

// EnsureRedisService makes sure the redis statefulset exists
func (r *RedisFailoverKubeClient) EnsureRedisService(rf *redisfailoverv1.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	svc := generateRedisService(rf, labels, ownerRefs)
	return r.K8SService.CreateIfNotExistsService(rf.Namespace, svc)
}

// EnsureNotPresentRedisService makes sure the redis service is not present
func (r *RedisFailoverKubeClient) EnsureNotPresentRedisService(rf *redisfailoverv1.RedisFailover) error {
	name := GetRedisName(rf)
	namespace := rf.Namespace
	// If the service exists (no get error), delete it
	if _, err := r.K8SService.GetService(namespace, name); err == nil {
		return r.K8SService.DeleteService(namespace, name)
	}
	return nil
}

// EnsureRedisStatefulset makes sure the pdb exists in the desired state
func (r *RedisFailoverKubeClient) ensurePodDisruptionBudget(rf *redisfailoverv1.RedisFailover, name string, component string, labels map[string]string, ownerRefs []metav1.OwnerReference) error {
	name = generateName(name, rf.Name)
	namespace := rf.Namespace

	minAvailable := intstr.FromInt(2)
	if rf.Spec.Redis.Replicas <= 2 {
		minAvailable = intstr.FromInt(int(rf.Spec.Redis.Replicas - 1))
	}

	labels = util.MergeLabels(labels, generateSelectorLabels(component, rf.Name))

	pdb := generatePodDisruptionBudget(name, namespace, labels, ownerRefs, minAvailable)

	return r.K8SService.CreateOrUpdatePodDisruptionBudget(namespace, pdb)
}
