package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// EndpointsRetrieve knows how to retrieve EP.
type EndpointsRetrieve struct {
	namespace string
	client    kubernetes.Interface
}

// NewEndpointsRetrieve returns a new EP retriever.
func NewEndpointsRetrieve(namespace string, client kubernetes.Interface) *PodRetrieve {
	return &EndpointsRetrieve{
		namespace: namespace,
		client:    client,
	}
}

// GetListerWatcher knows how to return a listerWatcher of a pod.
func (p *EndpointsRetrieve) GetListerWatcher() cache.ListerWatcher {

	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return p.client.CoreV1().Endpoints(p.namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return p.client.CoreV1().Endpoints(p.namespace).Watch(options)
		},
	}
}

// GetObject returns the empty pod.
func (p *EndpointsRetrieve) GetObject() runtime.Object {
	return &corev1.Pod{}
}
