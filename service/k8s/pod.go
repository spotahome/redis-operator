package k8s

import (
	"context"
	"encoding/json"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotahome/redis-operator/log"
)

// Pod the ServiceAccount service that knows how to interact with k8s to manage them
type Pod interface {
	GetPod(namespace string, name string) (*corev1.Pod, error)
	CreatePod(namespace string, pod *corev1.Pod) error
	UpdatePod(namespace string, pod *corev1.Pod) error
	CreateOrUpdatePod(namespace string, pod *corev1.Pod) error
	DeletePod(namespace string, name string) error
	ListPods(namespace string) (*corev1.PodList, error)
	UpdatePodLabels(namespace, podName string, labels map[string]string) error
}

// PodService is the pod service implementation using API calls to kubernetes.
type PodService struct {
	kubeClient kubernetes.Interface
	logger     log.Logger
}

// NewPodService returns a new Pod KubeService.
func NewPodService(kubeClient kubernetes.Interface, logger log.Logger) *PodService {
	logger = logger.With("service", "k8s.pod")
	return &PodService{
		kubeClient: kubeClient,
		logger:     logger,
	}
}

func (p *PodService) GetPod(namespace string, name string) (*corev1.Pod, error) {
	pod, err := p.kubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, err
}

func (p *PodService) CreatePod(namespace string, pod *corev1.Pod) error {
	_, err := p.kubeClient.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("pod", pod.Name).Infof("pod created")
	return nil
}
func (p *PodService) UpdatePod(namespace string, pod *corev1.Pod) error {
	_, err := p.kubeClient.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	p.logger.WithField("namespace", namespace).WithField("pod", pod.Name).Infof("pod updated")
	return nil
}
func (p *PodService) CreateOrUpdatePod(namespace string, pod *corev1.Pod) error {
	storedPod, err := p.GetPod(namespace, pod.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return p.CreatePod(namespace, pod)
		}
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.
	pod.ResourceVersion = storedPod.ResourceVersion
	return p.UpdatePod(namespace, pod)
}

func (p *PodService) DeletePod(namespace string, name string) error {
	return p.kubeClient.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (p *PodService) ListPods(namespace string) (*corev1.PodList, error) {
	return p.kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
}

//PatchStringValue specifies a patch operation for a string.
type PatchStringValue struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func (p *PodService) UpdatePodLabels(namespace, podName string, labels map[string]string) error {
	p.logger.Infof("Update pod label, namespace: %s, pod name: %s, labels: %v", namespace, podName, labels)

	var payloads []interface{}
	for labelKey, labelValue := range labels {
		payload := PatchStringValue{
			Op:    "replace",
			Path:  "/metadata/labels/" + labelKey,
			Value: labelValue,
		}
		payloads = append(payloads, payload)
	}
	payloadBytes, _ := json.Marshal(payloads)

	_, err := p.kubeClient.CoreV1().Pods(namespace).Patch(context.TODO(), podName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		p.logger.Errorf("Update pod labels failed, namespace: %s, pod name: %s, error: %v", namespace, podName, err)
	}
	return err
}
