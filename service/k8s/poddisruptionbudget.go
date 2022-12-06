package k8s

import (
	"context"

	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
)

// PodDisruptionBudget the ServiceAccount service that knows how to interact with k8s to manage them
type PodDisruptionBudget interface {
	GetPodDisruptionBudget(namespace string, name string) (*policyv1.PodDisruptionBudget, error)
	CreatePodDisruptionBudget(namespace string, podDisruptionBudget *policyv1.PodDisruptionBudget) error
	UpdatePodDisruptionBudget(namespace string, podDisruptionBudget *policyv1.PodDisruptionBudget) error
	CreateOrUpdatePodDisruptionBudget(namespace string, podDisruptionBudget *policyv1.PodDisruptionBudget) error
	DeletePodDisruptionBudget(namespace string, name string) error
}

// PodDisruptionBudgetService is the podDisruptionBudget service implementation using API calls to kubernetes.
type PodDisruptionBudgetService struct {
	kubeClient      kubernetes.Interface
	logger          log.Logger
	metricsRecorder metrics.Recorder
}

// NewPodDisruptionBudgetService returns a new PodDisruptionBudget KubeService.
func NewPodDisruptionBudgetService(kubeClient kubernetes.Interface, logger log.Logger, metricsRecorder metrics.Recorder) *PodDisruptionBudgetService {
	logger = logger.With("service", "k8s.podDisruptionBudget")
	return &PodDisruptionBudgetService{
		kubeClient:      kubeClient,
		logger:          logger,
		metricsRecorder: metricsRecorder,
	}
}

func (p *PodDisruptionBudgetService) GetPodDisruptionBudget(namespace string, name string) (*policyv1.PodDisruptionBudget, error) {
	podDisruptionBudget, err := p.kubeClient.PolicyV1().PodDisruptionBudgets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	recordMetrics(namespace, "PodDisruptionBudget", name, "GET", err, p.metricsRecorder)
	if err != nil {
		return nil, err
	}
	return podDisruptionBudget, nil
}

func (p *PodDisruptionBudgetService) CreatePodDisruptionBudget(namespace string, podDisruptionBudget *policyv1.PodDisruptionBudget) error {
	_, err := p.kubeClient.PolicyV1().PodDisruptionBudgets(namespace).Create(context.TODO(), podDisruptionBudget, metav1.CreateOptions{})
	recordMetrics(namespace, "PodDisruptionBudget", podDisruptionBudget.GetName(), "CREATE", err, p.metricsRecorder)
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("podDisruptionBudget", podDisruptionBudget.Name).Debugf("podDisruptionBudget created")
	return nil
}

func (p *PodDisruptionBudgetService) UpdatePodDisruptionBudget(namespace string, podDisruptionBudget *policyv1.PodDisruptionBudget) error {
	_, err := p.kubeClient.PolicyV1().PodDisruptionBudgets(namespace).Update(context.TODO(), podDisruptionBudget, metav1.UpdateOptions{})
	recordMetrics(namespace, "PodDisruptionBudget", podDisruptionBudget.GetName(), "UPDATE", err, p.metricsRecorder)
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("podDisruptionBudget", podDisruptionBudget.Name).Debugf("podDisruptionBudget updated")
	return nil
}

func (p *PodDisruptionBudgetService) CreateOrUpdatePodDisruptionBudget(namespace string, podDisruptionBudget *policyv1.PodDisruptionBudget) error {
	storedPodDisruptionBudget, err := p.GetPodDisruptionBudget(namespace, podDisruptionBudget.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return p.CreatePodDisruptionBudget(namespace, podDisruptionBudget)
		}
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.
	podDisruptionBudget.ResourceVersion = storedPodDisruptionBudget.ResourceVersion
	return p.UpdatePodDisruptionBudget(namespace, podDisruptionBudget)
}

func (p *PodDisruptionBudgetService) DeletePodDisruptionBudget(namespace string, name string) error {
	err := p.kubeClient.PolicyV1().PodDisruptionBudgets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	recordMetrics(namespace, "PodDisruptionBudget", name, "DELETE", err, p.metricsRecorder)
	return err
}
