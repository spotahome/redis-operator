package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
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
	kubeClient      kubernetes.Interface
	logger          log.Logger
	metricsRecorder metrics.Recorder
}

// NewServiceService returns a new Service KubeService.
func NewServiceService(kubeClient kubernetes.Interface, logger log.Logger, metricsRecorder metrics.Recorder) *ServiceService {
	logger = logger.With("service", "k8s.service")
	return &ServiceService{
		kubeClient:      kubeClient,
		logger:          logger,
		metricsRecorder: metricsRecorder,
	}
}

func (s *ServiceService) GetService(namespace string, name string) (*corev1.Service, error) {
	service, err := s.kubeClient.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	recordMetrics(namespace, "Service", name, "GET", err, s.metricsRecorder)
	if err != nil {
		return nil, err
	}
	return service, err
}

func (s *ServiceService) CreateService(namespace string, service *corev1.Service) error {
	_, err := s.kubeClient.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	recordMetrics(namespace, "Service", service.GetName(), "CREATE", err, s.metricsRecorder)
	if err != nil {
		return err
	}
	s.logger.WithField("namespace", namespace).WithField("serviceName", service.Name).Debugf("service created")
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
	_, err := s.kubeClient.CoreV1().Services(namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	recordMetrics(namespace, "Service", service.GetName(), "UPDATE", err, s.metricsRecorder)
	if err != nil {
		return err
	}
	s.logger.WithField("namespace", namespace).WithField("serviceName", service.Name).Debugf("service updated")
	return nil
}
func (s *ServiceService) CreateOrUpdateService(namespace string, service *corev1.Service) error {
	storedService, err := s.GetService(namespace, service.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return s.CreateService(namespace, service)
		}
		log.Errorf("Error while updating service %v in %v namespace : %v", service.GetName(), namespace, err)
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
	err := s.kubeClient.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	recordMetrics(namespace, "Service", name, "DELETE", err, s.metricsRecorder)
	return err
}

func (s *ServiceService) ListServices(namespace string) (*corev1.ServiceList, error) {
	serviceList, err := s.kubeClient.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	recordMetrics(namespace, "Service", metrics.NOT_APPLICABLE, "LIST", err, s.metricsRecorder)
	return serviceList, err
}
