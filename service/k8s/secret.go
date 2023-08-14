package k8s

import (
	"context"
	"fmt"

	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// Secret interacts with k8s to get secrets
type Secret interface {
	GetSecret(namespace, name string) (*corev1.Secret, error)
}

// SecretService is the secret service implementation using API calls to kubernetes.
type SecretService struct {
	kubeClient      kubernetes.Interface
	logger          log.Logger
	cacheStore      *cache.Store
	metricsRecorder metrics.Recorder
}

func NewSecretService(kubeClient kubernetes.Interface, logger log.Logger, metricsRecorder metrics.Recorder) *SecretService {

	logger = logger.With("service", "k8s.secret")
	rc := kubeClient.CoreV1().RESTClient().(*rest.RESTClient)
	var cacheStore *cache.Store
	var err error
	if ShouldUseCache() {
		cacheStore, err = SecretCacheStoreFromKubeClient(rc)
		if err != nil {
			logger.Errorf("unable to initialize cache: %v", err)
		}
	}
	return &SecretService{
		kubeClient:      kubeClient,
		logger:          logger,
		cacheStore:      cacheStore,
		metricsRecorder: metricsRecorder,
	}
}

func (s *SecretService) GetSecret(namespace, name string) (*corev1.Secret, error) {
	var secret *corev1.Secret
	var err error
	var exists bool
	if s.cacheStore != nil {
		c := *s.cacheStore
		var item interface{}
		item, exists, err = c.GetByKey(fmt.Sprintf("%v/%v", namespace, name))
		if exists && nil == err {
			secret = item.(*corev1.Secret)
		}
		if !exists {
			err = fmt.Errorf("secret %v not found in namespace %v", name, namespace)
		}
	} else {
		secret, err = s.kubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	}

	recordMetrics(namespace, "Secret", name, "GET", err, s.metricsRecorder)

	return secret, err
}
