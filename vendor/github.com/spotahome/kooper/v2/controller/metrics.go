package controller

import (
	"context"
	"time"
)

// MetricsRecorder knows how to record metrics of a controller.
type MetricsRecorder interface {
	// IncResourceEvent increments in one the metric records of a queued event.
	IncResourceEventQueued(ctx context.Context, controller string, isRequeue bool)
	// ObserveResourceInQueueDuration measures how long takes to dequeue a queued object. If the object is already in queue
	// it will be measured once, since the first time it was added to the queue.
	ObserveResourceInQueueDuration(ctx context.Context, controller string, queuedAt time.Time)
	// ObserveResourceProcessingDuration measures how long it takes to process a resources (handling).
	ObserveResourceProcessingDuration(ctx context.Context, controller string, success bool, startProcessingAt time.Time)
	// RegisterResourceQueueLengthFunc will register a function that will be called
	// by the metrics recorder to get the length of a queue at a given point in time.
	RegisterResourceQueueLengthFunc(controller string, f func(context.Context) int) error
}

// DummyMetricsRecorder is a dummy metrics recorder.
var DummyMetricsRecorder = dummy(0)
var _ MetricsRecorder = DummyMetricsRecorder

type dummy int

func (dummy) IncResourceEventQueued(context.Context, string, bool)                       {}
func (dummy) ObserveResourceInQueueDuration(context.Context, string, time.Time)          {}
func (dummy) ObserveResourceProcessingDuration(context.Context, string, bool, time.Time) {}
func (dummy) RegisterResourceQueueLengthFunc(controller string, f func(context.Context) int) error {
	return nil
}
