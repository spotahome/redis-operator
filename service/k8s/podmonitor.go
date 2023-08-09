package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
)

// PodMonitor the ServiceAccount service that knows how to interact with k8s to manage them
type PodMonitor interface {
	GetPodMonitor(namespace string, name string) (*prometheusv1.PodMonitor, error)
	CreatePodMonitor(namespace string, podMonitor *prometheusv1.PodMonitor) error
	CreateIfNotExistsPodMonitor(namespace string, podMonitor *prometheusv1.PodMonitor) error
	UpdatePodMonitor(namespace string, podMonitor *prometheusv1.PodMonitor) error
	CreateOrUpdatePodMonitor(namespace string, podMonitor *prometheusv1.PodMonitor) error
	DeletePodMonitor(namespace string, name string) error
	ListPodMonitor(namespace string) (*prometheusv1.PodMonitorList, error)
}

// PodMonitorService is the podMonitor service implementation using API calls to kubernetes.
type PodMonitorService struct {
	monitoringV1Client monitoringv1.MonitoringV1Interface
	logger             log.Logger
	metricsRecorder    metrics.Recorder
}

// NewPodMonitorService returns a new Service KubeService.
func NewPodMonitorService(monitoringV1Client monitoringv1.MonitoringV1Interface, logger log.Logger, metricsRecorder metrics.Recorder) *PodMonitorService {
	logger = logger.With("service", "k8s.podMonitor") // ??? k8s.podMonitor or prometheus.podMonitor
	return &PodMonitorService{
		monitoringV1Client: monitoringV1Client,
		logger:             logger,
		metricsRecorder:    metricsRecorder,
	}
}

func (p *PodMonitorService) GetPodMonitor(namespace string, name string) (*prometheusv1.PodMonitor, error) {
	podMonitor, err := p.monitoringV1Client.PodMonitors(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	recordMetrics(namespace, "PodMonitor", name, "GET", err, p.metricsRecorder)
	if err != nil {
		return nil, err
	}
	return podMonitor, err
}

func (p *PodMonitorService) CreatePodMonitor(namespace string, podMonitor *prometheusv1.PodMonitor) error {
	_, err := p.monitoringV1Client.PodMonitors(namespace).Create(context.TODO(), podMonitor, metav1.CreateOptions{})
	recordMetrics(namespace, "PodMonitor", podMonitor.GetName(), "CREATE", err, p.metricsRecorder)
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("podMonitorName", podMonitor.Name).Debugf("podMonitor created")
	return nil
}

func (p *PodMonitorService) CreateIfNotExistsPodMonitor(namespace string, podMonitor *prometheusv1.PodMonitor) error {
	if _, err := p.GetPodMonitor(namespace, podMonitor.Name); err != nil {
		if errors.IsNotFound(err) {
			return p.CreatePodMonitor(namespace, podMonitor)
		}
		return err
	}
	return nil
}

func (p *PodMonitorService) UpdatePodMonitor(namespace string, podMonitor *prometheusv1.PodMonitor) error {
	_, err := p.monitoringV1Client.PodMonitors(namespace).Update(context.TODO(), podMonitor, metav1.UpdateOptions{})
	recordMetrics(namespace, "PodMonitor", podMonitor.GetName(), "UPDATE", err, p.metricsRecorder)
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("podMonitorName", podMonitor.Name).Debugf("podMonitor updated")
	return nil
}
func (p *PodMonitorService) CreateOrUpdatePodMonitor(namespace string, podMonitor *prometheusv1.PodMonitor) error {
	storedPodMonitor, err := p.GetPodMonitor(namespace, podMonitor.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return p.CreatePodMonitor(namespace, podMonitor)
		}
		log.Errorf("Error while updating podMonitor %v in %v namespace : %v", podMonitor.GetName(), namespace, err)
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.
	podMonitor.ResourceVersion = storedPodMonitor.ResourceVersion
	return p.UpdatePodMonitor(namespace, podMonitor)
}

func (p *PodMonitorService) DeletePodMonitor(namespace string, name string) error {
	propagation := metav1.DeletePropagationForeground
	err := p.monitoringV1Client.PodMonitors(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	recordMetrics(namespace, "PodMonitor", name, "DELETE", err, p.metricsRecorder)
	return err
}

func (p *PodMonitorService) ListPodMonitor(namespace string) (*prometheusv1.PodMonitorList, error) {
	podMonitorList, err := p.monitoringV1Client.PodMonitors(namespace).List(context.TODO(), metav1.ListOptions{})
	recordMetrics(namespace, "PodMonitor", metrics.NOT_APPLICABLE, "LIST", err, p.metricsRecorder)
	return podMonitorList, err
}
