package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// Handler knows how to handle the received resources from a kubernetes cluster.
type Handler interface {
	Handle(context.Context, runtime.Object) error
}

//go:generate mockery -case underscore -output controllermock -outpkg controllermock -name Handler

// HandlerFunc knows how to handle resource adds.
type HandlerFunc func(context.Context, runtime.Object) error

// Handle satisfies controller.Handler interface.
func (h HandlerFunc) Handle(ctx context.Context, obj runtime.Object) error {
	if h == nil {
		return fmt.Errorf("handle func is required")
	}
	return h(ctx, obj)
}
