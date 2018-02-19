package handler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// Handler knows how to handle the received resources from a kubernetes cluster.
type Handler interface {
	Add(obj runtime.Object) error
	Delete(string) error
}

// AddFunc knows how to handle resource adds.
type AddFunc func(obj runtime.Object) error

// DeleteFunc knows how to handle resource deletes.
type DeleteFunc func(string) error

// HandlerFunc is a handler that is created from functions that the
// Handler interface requires.
type HandlerFunc struct {
	AddFunc    AddFunc
	DeleteFunc DeleteFunc
}

// Add satisfies Handler interface.
func (h *HandlerFunc) Add(obj runtime.Object) error {
	if h.AddFunc == nil {
		return fmt.Errorf("function can't be nil")
	}
	return h.AddFunc(obj)
}

// Delete satisfies Handler interface.
func (h *HandlerFunc) Delete(s string) error {
	if h.DeleteFunc == nil {
		return fmt.Errorf("function can't be nil")
	}
	return h.DeleteFunc(s)
}
