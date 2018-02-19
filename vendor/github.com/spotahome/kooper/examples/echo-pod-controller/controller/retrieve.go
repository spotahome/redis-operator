package controller

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// PodRetrieve knows how to retrieve pods.
type PodRetrieve struct {
	namespace string
	client    kubernetes.Interface
}

// NewPodRetrieve returns a new pod retriever.
func NewPodRetrieve(namespace string, client kubernetes.Interface) *PodRetrieve {
	return &PodRetrieve{
		namespace: namespace,
		client:    client,
	}
}

// GetListerWatcher knows how to return a listerWatcher of a pod.
func (p *PodRetrieve) GetListerWatcher() cache.ListerWatcher {

	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return p.client.CoreV1().Pods(p.namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return p.client.CoreV1().Pods(p.namespace).Watch(options)
		},
	}
}

// GetObject returns the empty pod.
func (p *PodRetrieve) GetObject() runtime.Object {
	return &corev1.Pod{}
}
