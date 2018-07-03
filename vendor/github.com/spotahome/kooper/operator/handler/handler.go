package handler

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// Handler knows how to handle the received resources from a kubernetes cluster.
type Handler interface {
	Add(context.Context, runtime.Object) error
	Delete(context.Context, string) error
}

// AddFunc knows how to handle resource adds.
type AddFunc func(context.Context, runtime.Object) error

// DeleteFunc knows how to handle resource deletes.
type DeleteFunc func(context.Context, string) error

// HandlerFunc is a handler that is created from functions that the
// Handler interface requires.
type HandlerFunc struct {
	AddFunc    AddFunc
	DeleteFunc DeleteFunc
}

// Add satisfies Handler interface.
func (h *HandlerFunc) Add(ctx context.Context, obj runtime.Object) error {
	if h.AddFunc == nil {
		return fmt.Errorf("function can't be nil")
	}
	return h.AddFunc(ctx, obj)
}

// Delete satisfies Handler interface.
func (h *HandlerFunc) Delete(ctx context.Context, s string) error {
	if h.DeleteFunc == nil {
		return fmt.Errorf("function can't be nil")
	}
	return h.DeleteFunc(ctx, s)
}
