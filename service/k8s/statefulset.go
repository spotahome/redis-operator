package k8s

import (
	"fmt"
	"strings"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotahome/redis-operator/log"
)

// StatefulSet the StatefulSet service that knows how to interact with k8s to manage them
type StatefulSet interface {
	GetStatefulSet(namespace, name string) (*appsv1beta2.StatefulSet, error)
	GetStatefulSetPods(namespace, name string) (*corev1.PodList, error)
	CreateStatefulSet(namespace string, statefulSet *appsv1beta2.StatefulSet) error
	UpdateStatefulSet(namespace string, statefulSet *appsv1beta2.StatefulSet) error
	CreateOrUpdateStatefulSet(namespace string, statefulSet *appsv1beta2.StatefulSet) error
	DeleteStatefulSet(namespace string, name string) error
	ListStatefulSets(namespace string) (*appsv1beta2.StatefulSetList, error)
}

// StatefulSetService is the service account service implementation using API calls to kubernetes.
type StatefulSetService struct {
	kubeClient kubernetes.Interface
	logger     log.Logger
}

// NewStatefulSetService returns a new StatefulSet KubeService.
func NewStatefulSetService(kubeClient kubernetes.Interface, logger log.Logger) *StatefulSetService {
	logger = logger.With("service", "k8s.statefulSet")
	return &StatefulSetService{
		kubeClient: kubeClient,
		logger:     logger,
	}
}

func (s *StatefulSetService) GetStatefulSet(namespace, name string) (*appsv1beta2.StatefulSet, error) {
	statefulSet, err := s.kubeClient.AppsV1beta2().StatefulSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return statefulSet, err
}

func (s *StatefulSetService) GetStatefulSetPods(namespace, name string) (*corev1.PodList, error) {
	statefulSet, err := s.GetStatefulSet(namespace, name)
	if err != nil {
		return nil, err
	}
	labels := []string{}
	for k, v := range statefulSet.Spec.Selector.MatchLabels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(labels, ",")
	return s.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector})
}

func (s *StatefulSetService) CreateStatefulSet(namespace string, statefulSet *appsv1beta2.StatefulSet) error {
	_, err := s.kubeClient.AppsV1beta2().StatefulSets(namespace).Create(statefulSet)
	if err != nil {
		return err
	}
	s.logger.WithField("namespace", namespace).WithField("statefulSet", statefulSet.ObjectMeta.Name).Infof("statefulSet created")
	return err
}

func (s *StatefulSetService) UpdateStatefulSet(namespace string, statefulSet *appsv1beta2.StatefulSet) error {
	_, err := s.kubeClient.AppsV1beta2().StatefulSets(namespace).Update(statefulSet)
	if err != nil {
		return err
	}
	s.logger.WithField("namespace", namespace).WithField("statefulSet", statefulSet.ObjectMeta.Name).Infof("statefulSet updated")
	return err
}

func (s *StatefulSetService) CreateOrUpdateStatefulSet(namespace string, statefulSet *appsv1beta2.StatefulSet) error {
	storedStatefulSet, err := s.GetStatefulSet(namespace, statefulSet.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return s.CreateStatefulSet(namespace, statefulSet)
		}
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.
	statefulSet.ResourceVersion = storedStatefulSet.ResourceVersion
	return s.UpdateStatefulSet(namespace, statefulSet)
}

func (s *StatefulSetService) DeleteStatefulSet(namespace, name string) error {
	propagation := metav1.DeletePropagationForeground
	return s.kubeClient.AppsV1beta2().StatefulSets(namespace).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &propagation})
}

func (s *StatefulSetService) ListStatefulSets(namespace string) (*appsv1beta2.StatefulSetList, error) {
	return s.kubeClient.AppsV1beta2().StatefulSets(namespace).List(metav1.ListOptions{})
}
