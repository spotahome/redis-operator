package retrieve

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Namespace knows how to retrieve namespaces.
type Namespace struct {
	lw  cache.ListerWatcher
	obj runtime.Object
}

// NewNamespace returns a Namespace retriever.
func NewNamespace(client kubernetes.Interface) *Namespace {
	return &Namespace{
		lw: &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Namespaces().List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Namespaces().Watch(options)
			},
		},
		obj: &corev1.Namespace{},
	}
}

// GetListerWatcher knows how to retreive Namespaces.
func (n *Namespace) GetListerWatcher() cache.ListerWatcher {
	return n.lw
}

// GetObject returns the namespace Object.
func (n *Namespace) GetObject() runtime.Object {
	return n.obj
}
