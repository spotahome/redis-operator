package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"

	np "k8s.io/api/networking/v1"
)

// NetworkPolicy the ServiceAccount service that knows how to interact with k8s to manage them
type NetworkPolicy interface {
	GetNetworkPolicy(namespace string, name string) (*np.NetworkPolicy, error)
	CreateNetworkPolicy(namespace string, networkPolicy *np.NetworkPolicy) error
	UpdateNetworkPolicy(namespace string, networkPolicy *np.NetworkPolicy) error
	CreateOrUpdateNetworkPolicy(namespace string, networkPolicy *np.NetworkPolicy) error
	DeleteNetworkPolicy(namespace string, name string) error
}

// NetworkPolicyService is the networkPolicy service implementation using API calls to kubernetes.
type NetworkPolicyService struct {
	kubeClient      kubernetes.Interface
	logger          log.Logger
	metricsRecorder metrics.Recorder
}

// NewNetworkPolicyService returns a new NetworkPolicy KubeService.
func NewNetworkPolicyService(kubeClient kubernetes.Interface, logger log.Logger, metricsRecorder metrics.Recorder) *NetworkPolicyService {
	logger = logger.With("service", "k8s.networkPolicy")
	return &NetworkPolicyService{
		kubeClient:      kubeClient,
		logger:          logger,
		metricsRecorder: metricsRecorder,
	}
}

func (p *NetworkPolicyService) GetNetworkPolicy(namespace string, name string) (*np.NetworkPolicy, error) {
	networkPolicy, err := p.kubeClient.NetworkingV1().NetworkPolicies(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	recordMetrics(namespace, "NetworkPolicy", name, "GET", err, p.metricsRecorder)
	if err != nil {
		return nil, err
	}
	return networkPolicy, err
}

func (p *NetworkPolicyService) CreateNetworkPolicy(namespace string, networkPolicy *np.NetworkPolicy) error {
	_, err := p.kubeClient.NetworkingV1().NetworkPolicies(namespace).Create(context.TODO(), networkPolicy, metav1.CreateOptions{})
	recordMetrics(namespace, "NetworkPolicy", networkPolicy.GetName(), "CREATE", err, p.metricsRecorder)
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("networkPolicy", networkPolicy.Name).Debugf("NetworkPolicy created")
	return nil
}
func (p *NetworkPolicyService) UpdateNetworkPolicy(namespace string, networkPolicy *np.NetworkPolicy) error {
	_, err := p.kubeClient.NetworkingV1().NetworkPolicies(namespace).Update(context.TODO(), networkPolicy, metav1.UpdateOptions{})
	recordMetrics(namespace, "NetworkPolicy", networkPolicy.GetName(), "UPDATE", err, p.metricsRecorder)
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("networkPolicy", networkPolicy.Name).Debugf("NetworkPolicy updated")
	return nil
}
func (p *NetworkPolicyService) CreateOrUpdateNetworkPolicy(namespace string, networkPolicy *np.NetworkPolicy) error {
	storedNetworkPolicy, err := p.GetNetworkPolicy(namespace, networkPolicy.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return p.CreateNetworkPolicy(namespace, networkPolicy)
		}
		log.Errorf("Error while updating service %v in %v namespace : %v", networkPolicy.GetName(), namespace, err)
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.
	networkPolicy.ResourceVersion = storedNetworkPolicy.ResourceVersion
	return p.UpdateNetworkPolicy(namespace, networkPolicy)
}

func (p *NetworkPolicyService) DeleteNetworkPolicy(namespace string, name string) error {
	propagation := metav1.DeletePropagationForeground
	err := p.kubeClient.NetworkingV1().NetworkPolicies(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	recordMetrics(namespace, "NetworkPolicy", name, "DELETE", err, p.metricsRecorder)
	return err
}

