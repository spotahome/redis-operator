package controller

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// Retriever is how a controller will retrieve the events on the resources from
// the APÎ server.
//
// A Retriever is bound to a single type.
type Retriever interface {
	List(ctx context.Context, options metav1.ListOptions) (runtime.Object, error)
	Watch(ctx context.Context, options metav1.ListOptions) (watch.Interface, error)
}

type listerWatcherRetriever struct {
	lw cache.ListerWatcher
}

// RetrieverFromListerWatcher returns a Retriever from a Kubernetes client-go cache.ListerWatcher.
// If the received lister watcher is nil it will error.
func RetrieverFromListerWatcher(lw cache.ListerWatcher) (Retriever, error) {
	if lw == nil {
		return nil, fmt.Errorf("listerWatcher can't be nil")
	}
	return listerWatcherRetriever{lw: lw}, nil
}

// MustRetrieverFromListerWatcher returns a Retriever from a Kubernetes client-go cache.ListerWatcher
// if there is an error it will panic.
func MustRetrieverFromListerWatcher(lw cache.ListerWatcher) Retriever {
	r, err := RetrieverFromListerWatcher(lw)
	if lw == nil {
		panic(err)
	}
	return r
}

func (l listerWatcherRetriever) List(_ context.Context, options metav1.ListOptions) (runtime.Object, error) {
	return l.lw.List(options)
}
func (l listerWatcherRetriever) Watch(_ context.Context, options metav1.ListOptions) (watch.Interface, error) {
	return l.lw.Watch(options)
}
