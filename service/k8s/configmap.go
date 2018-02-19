package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotahome/redis-operator/log"
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
	kubeClient kubernetes.Interface
	logger     log.Logger
}

// NewConfigMapService returns a new ConfigMap KubeService.
func NewConfigMapService(kubeClient kubernetes.Interface, logger log.Logger) *ConfigMapService {
	logger = logger.With("service", "k8s.configMap")
	return &ConfigMapService{
		kubeClient: kubeClient,
		logger:     logger,
	}
}

func (p *ConfigMapService) GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error) {
	configMap, err := p.kubeClient.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return configMap, err
}

func (p *ConfigMapService) CreateConfigMap(namespace string, configMap *corev1.ConfigMap) error {
	_, err := p.kubeClient.CoreV1().ConfigMaps(namespace).Create(configMap)
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("configMap", configMap.Name).Infof("configMap created")
	return nil
}
func (p *ConfigMapService) UpdateConfigMap(namespace string, configMap *corev1.ConfigMap) error {
	_, err := p.kubeClient.CoreV1().ConfigMaps(namespace).Update(configMap)
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("configMap", configMap.Name).Infof("configMap updated")
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
	return p.kubeClient.CoreV1().ConfigMaps(namespace).Delete(name, &metav1.DeleteOptions{})
}

func (p *ConfigMapService) ListConfigMaps(namespace string) (*corev1.ConfigMapList, error) {
	return p.kubeClient.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{})
}
