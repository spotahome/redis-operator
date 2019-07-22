package k8s

import (
	"github.com/spotahome/redis-operator/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Endpoints interface {
	ListEndpoints(namespace string) (*corev1.EndpointsList, error)
}

// EndpointsService is the service account service implementation using API calls to kubernetes.
type EnpointsService struct {
	kubeClient kubernetes.Interface
	logger     log.Logger
}

func NewEndpointsService(kubeClient kubernetes.Interface, logger log.Logger) *EnpointsService {
	logger = logger.With("service", "k8s.endpoints")
	return &EnpointsService{
		kubeClient: kubeClient,
		logger:     logger,
	}
}

func (e *EnpointsService) ListEndpoints(namespace string) (*corev1.EndpointsList, error) {
	//FieldSelector: "metadata.name=" + GetSentinelName(rf)
	return e.kubeClient.CoreV1().Endpoints(namespace).List(metav1.ListOptions{})
}
