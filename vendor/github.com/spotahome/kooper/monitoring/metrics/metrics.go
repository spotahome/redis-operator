package metrics

import "time"

// EventType is the event type handled by the controller.
type EventType string

const (
	//AddEvent is the add event.
	AddEvent EventType = "add"
	// DeleteEvent is the delete event.
	DeleteEvent EventType = "delete"
	// RequeueEvent is a requeued event (unknown state when handling again).
	RequeueEvent EventType = "requeue"
)

// Recorder knows how to record metrics all over the application.
type Recorder interface {
	// IncResourceEvent increments in one the metric records of a queued event.
	IncResourceEventQueued(controller string, eventType EventType)
	// IncResourceEventProcessed increments in one the metric records processed event.
	IncResourceEventProcessed(controller string, eventType EventType)
	// IncResourceEventProcessedError increments in one the metric records of a processed event in error.
	IncResourceEventProcessedError(controller string, eventType EventType)
	// ObserveDurationResourceEventProcessed measures the duration it took to process a event.
	ObserveDurationResourceEventProcessed(controller string, eventType EventType, start time.Time)
}
