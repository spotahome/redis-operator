package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	"k8s.io/client-go/tools/cache"
)

// ConfigMap the ServiceAccount service that knows how to interact with k8s to manage them
type ConfigMap interface {
	GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error)
	CreateConfigMap(namespace string, configMap *corev1.ConfigMap) error
	UpdateConfigMap(namespace string, configMap *corev1.ConfigMap) error
	CreateOrUpdateConfigMap(namespace string, np *corev1.ConfigMap) error
	DeleteConfigMap(namespace string, name string) error
	ListConfigMaps(namespace string) (*corev1.ConfigMapList, error)
}

// ConfigMapService is the configMap service implementation using API calls to kubernetes.
type ConfigMapService struct {
	kubeClient      kubernetes.Interface
	logger          log.Logger
	cacheStore      *cache.Store
	metricsRecorder metrics.Recorder
}

// NewConfigMapService returns a new ConfigMap KubeService.
func NewConfigMapService(kubeClient kubernetes.Interface, logger log.Logger, metricsRecorder metrics.Recorder) *ConfigMapService {
	logger = logger.With("service", "k8s.configMap")
	var err error
	rc := kubeClient.CoreV1().RESTClient().(*rest.RESTClient)
	var cmCacheStore *cache.Store
	if ShouldUseCache() {
		cmCacheStore, err = ConfigMapCacheStoreFromKubeClient(rc)
		if err != nil {
			logger.Errorf("unable to initialize cache: %v", err)
		}
	}
	return &ConfigMapService{
		kubeClient:      kubeClient,
		logger:          logger,
		cacheStore:      cmCacheStore,
		metricsRecorder: metricsRecorder,
	}
}

func (p *ConfigMapService) GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error) {
	var cm *corev1.ConfigMap
	var err error
	var exists bool
	if p.cacheStore != nil {
		c := *p.cacheStore
		var item interface{}
		item, exists, err = c.GetByKey(fmt.Sprintf("%v/%v", namespace, name))
		if exists && nil == err {
			cm = item.(*corev1.ConfigMap)
		}
		if !exists {
			err = fmt.Errorf("configmap %v not found in namespace %v", name, namespace)
		}
	} else {
		cm, err = p.kubeClient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	}

	recordMetrics(namespace, "ConfigMap", name, "GET", err, p.metricsRecorder)

	return cm, err
}

func (p *ConfigMapService) CreateConfigMap(namespace string, configMap *corev1.ConfigMap) error {
	_, err := p.kubeClient.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	recordMetrics(namespace, "ConfigMap", configMap.GetName(), "CREATE", err, p.metricsRecorder)
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("configMap", configMap.Name).Debugf("configMap created")
	return nil
}
func (p *ConfigMapService) UpdateConfigMap(namespace string, configMap *corev1.ConfigMap) error {
	_, err := p.kubeClient.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	recordMetrics(namespace, "ConfigMap", configMap.GetName(), "UPDATE", err, p.metricsRecorder)
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("configMap", configMap.Name).Debugf("configMap updated")
	return nil
}
func (p *ConfigMapService) CreateOrUpdateConfigMap(namespace string, configMap *corev1.ConfigMap) error {
	storedConfigMap, err := p.GetConfigMap(namespace, configMap.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return p.CreateConfigMap(namespace, configMap)
		}
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.
	configMap.ResourceVersion = storedConfigMap.ResourceVersion
	return p.UpdateConfigMap(namespace, configMap)
}

func (p *ConfigMapService) DeleteConfigMap(namespace string, name string) error {
	err := p.kubeClient.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	recordMetrics(namespace, "ConfigMap", name, "DELETE", err, p.metricsRecorder)
	return err
}

func (p *ConfigMapService) ListConfigMaps(namespace string) (*corev1.ConfigMapList, error) {
	objects, err := p.kubeClient.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	recordMetrics(namespace, "ConfigMap", metrics.NOT_APPLICABLE, "LIST", err, p.metricsRecorder)
	return objects, err
}
