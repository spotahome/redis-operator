package retrieve

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

// Retriever is a way of wrapping  kubernetes lister watchers so they are easy to pass & manage them
// on Controllers.
type Retriever interface {
	GetListerWatcher() cache.ListerWatcher
	GetObject() runtime.Object
}

// Resource is a helper so you can don't need to create a new type of the
// Retriever interface.
type Resource struct {
	ListerWatcher cache.ListerWatcher
	Object        runtime.Object
}

// GetListerWatcher satisfies retriever interface.
func (r *Resource) GetListerWatcher() cache.ListerWatcher {
	return r.ListerWatcher
}

// GetObject satisfies retriever interface
func (r *Resource) GetObject() runtime.Object {
	return r.Object
}
