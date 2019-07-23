package k8s

import (
	"github.com/spotahome/redis-operator/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// Endpoints Endpoints
type Endpoints interface {
	ListEndpoints(namespace string, opts metav1.ListOptions) (*corev1.EndpointsList, error)
	WatchEndpoints(namespace string, opts metav1.ListOptions) (watch.Interface, error)
}

// EnpointsService is the service account service implementation using API calls to kubernetes.
type EnpointsService struct {
	kubeClient kubernetes.Interface
	logger     log.Logger
}

// NewEndpointsService NewEndpointsService
func NewEndpointsService(kubeClient kubernetes.Interface, logger log.Logger) *EnpointsService {
	logger = logger.With("service", "k8s.endpoints")
	return &EnpointsService{
		kubeClient: kubeClient,
		logger:     logger,
	}
}

// ListEndpoints ListEndpoints
func (e *EnpointsService) ListEndpoints(namespace string, opts metav1.ListOptions) (*corev1.EndpointsList, error) {
	// opts := metav1.ListOptions{FieldSelector: "metadata.name=" + name}
	return e.kubeClient.CoreV1().Endpoints(namespace).List(opts)
}

// WatchEndpoints WatchEndpoints
func (e *EnpointsService) WatchEndpoints(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return e.kubeClient.CoreV1().Endpoints(namespace).Watch(opts)
}
