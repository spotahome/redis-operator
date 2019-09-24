package k8s

import (
	"github.com/spotahome/redis-operator/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Secret interacts with k8s to get secrets
type Secret interface {
	GetSecret(namespace, name string) (*corev1.Secret, error)
}

// SecretService is the secret service implementation using API calls to kubernetes.
type SecretService struct {
	kubeClient kubernetes.Interface
	logger     log.Logger
}

func NewSecretService(kubeClient kubernetes.Interface, logger log.Logger) *SecretService {

	logger = logger.With("service", "k8s.secret")
	return &SecretService{
		kubeClient: kubeClient,
		logger:     logger,
	}
}

func (s *SecretService) GetSecret(namespace, name string) (*corev1.Secret, error) {

	secret, err := s.kubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, err
}
