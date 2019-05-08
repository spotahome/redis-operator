package k8s

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotahome/redis-operator/log"
)

// Deployment the Deployment service that knows how to interact with k8s to manage them
type Deployment interface {
	GetDeployment(namespace, name string) (*appsv1.Deployment, error)
	GetDeploymentPods(namespace, name string) (*corev1.PodList, error)
	CreateDeployment(namespace string, deployment *appsv1.Deployment) error
	UpdateDeployment(namespace string, deployment *appsv1.Deployment) error
	CreateOrUpdateDeployment(namespace string, deployment *appsv1.Deployment) error
	DeleteDeployment(namespace string, name string) error
	ListDeployments(namespace string) (*appsv1.DeploymentList, error)
}

// DeploymentService is the service account service implementation using API calls to kubernetes.
type DeploymentService struct {
	kubeClient kubernetes.Interface
	logger     log.Logger
}

// NewDeploymentService returns a new Deployment KubeService.
func NewDeploymentService(kubeClient kubernetes.Interface, logger log.Logger) *DeploymentService {
	logger = logger.With("service", "k8s.deployment")
	return &DeploymentService{
		kubeClient: kubeClient,
		logger:     logger,
	}
}

// GetDeployment will retrieve the requested deployment based on namespace and name
func (d *DeploymentService) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	deployment, err := d.kubeClient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return deployment, err
}

// GetDeploymentPods will retrieve the pods managed by a given deployment
func (d *DeploymentService) GetDeploymentPods(namespace, name string) (*corev1.PodList, error) {
	deployment, err := d.kubeClient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	labels := []string{}
	for k, v := range deployment.Spec.Selector.MatchLabels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(labels, ",")
	return d.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector})
}

// CreateDeployment will create the given deployment
func (d *DeploymentService) CreateDeployment(namespace string, deployment *appsv1.Deployment) error {
	_, err := d.kubeClient.AppsV1().Deployments(namespace).Create(deployment)
	if err != nil {
		return err
	}
	d.logger.WithField("namespace", namespace).WithField("deployment", deployment.ObjectMeta.Name).Infof("deployment created")
	return err
}

// UpdateDeployment will update the given deployment
func (d *DeploymentService) UpdateDeployment(namespace string, deployment *appsv1.Deployment) error {
	_, err := d.kubeClient.AppsV1().Deployments(namespace).Update(deployment)
	if err != nil {
		return err
	}
	d.logger.WithField("namespace", namespace).WithField("deployment", deployment.ObjectMeta.Name).Infof("deployment updated")
	return err
}

// CreateOrUpdateDeployment will update the given deployment or create it if does not exist
func (d *DeploymentService) CreateOrUpdateDeployment(namespace string, deployment *appsv1.Deployment) error {
	storedDeployment, err := d.GetDeployment(namespace, deployment.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return d.CreateDeployment(namespace, deployment)
		}
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.
	deployment.ResourceVersion = storedDeployment.ResourceVersion
	return d.UpdateDeployment(namespace, deployment)
}

// DeleteDeployment will delete the given deployment
func (d *DeploymentService) DeleteDeployment(namespace, name string) error {
	propagation := metav1.DeletePropagationForeground
	return d.kubeClient.AppsV1().Deployments(namespace).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &propagation})
}

// ListDeployments will give all the deployments on a given namespace
func (d *DeploymentService) ListDeployments(namespace string) (*appsv1.DeploymentList, error) {
	return d.kubeClient.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
}
