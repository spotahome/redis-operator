package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotahome/redis-operator/log"
)

// Service the ServiceAccount service that knows how to interact with k8s to manage them
type Service interface {
	GetService(namespace string, name string) (*corev1.Service, error)
	CreateService(namespace string, service *corev1.Service) error
	CreateIfNotExistsService(namespace string, service *corev1.Service) error
	UpdateService(namespace string, service *corev1.Service) error
	CreateOrUpdateService(namespace string, service *corev1.Service) error
	DeleteService(namespace string, name string) error
	ListServices(namespace string) (*corev1.ServiceList, error)
}

// ServiceService is the service service implementation using API calls to kubernetes.
type ServiceService struct {
	kubeClient kubernetes.Interface
	logger     log.Logger
}

// NewServiceService returns a new Service KubeService.
func NewServiceService(kubeClient kubernetes.Interface, logger log.Logger) *ServiceService {
	logger = logger.With("service", "k8s.service")
	return &ServiceService{
		kubeClient: kubeClient,
		logger:     logger,
	}
}

func (s *ServiceService) GetService(namespace string, name string) (*corev1.Service, error) {
	service, err := s.kubeClient.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return service, err
}

func (s *ServiceService) CreateService(namespace string, service *corev1.Service) error {
	_, err := s.kubeClient.CoreV1().Services(namespace).Create(service)
	if err != nil {
		return err
	}
	s.logger.WithField("namespace", namespace).WithField("serviceName", service.Name).Infof("service created")
	return nil
}

func (s *ServiceService) CreateIfNotExistsService(namespace string, service *corev1.Service) error {
	if _, err := s.GetService(namespace, service.Name); err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return s.CreateService(namespace, service)
		}
		return err
	}
	return nil
}

func (s *ServiceService) UpdateService(namespace string, service *corev1.Service) error {
	_, err := s.kubeClient.CoreV1().Services(namespace).Update(service)
	if err != nil {
		return err
	}
	s.logger.WithField("namespace", namespace).WithField("serviceName", service.Name).Infof("service updated")
	return nil
}
func (s *ServiceService) CreateOrUpdateService(namespace string, service *corev1.Service) error {
	storedService, err := s.GetService(namespace, service.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return s.CreateService(namespace, service)
		}
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.
	service.ResourceVersion = storedService.ResourceVersion
	return s.UpdateService(namespace, service)
}

func (s *ServiceService) DeleteService(namespace string, name string) error {
	propagation := metav1.DeletePropagationForeground
	return s.kubeClient.CoreV1().Services(namespace).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &propagation})
}

func (s *ServiceService) ListServices(namespace string) (*corev1.ServiceList, error) {
	return s.kubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{})
}
