package metrics

import "time"

// Recorder knows how to record metrics all over the application.
type Recorder interface {
	// IncResourceAddEvent increments in one the metric records of a queued add
	// event in a resource.
	IncResourceAddEventQueued(handler string)
	// IncResourceDeleteEvent increments in one the metric records of a queued delete
	// event in a resource.
	IncResourceDeleteEventQueued(handler string)
	// IncResourceAddEventProcessedSuccess increments in one the metric records of a
	// processed add event in success.
	IncResourceAddEventProcessedSuccess(handler string)
	// IncResourceAddEventProcessedError increments in one the metric records of a
	// processed add event in error.
	IncResourceAddEventProcessedError(handler string)
	// IncResourceDeleteEventProcessedSuccess increments in one the metric records of a
	// processed deleteevent in success.
	IncResourceDeleteEventProcessedSuccess(handler string)
	// IncResourceDeleteEventProcessedError increments in one the metric records of a
	// processed delete event in error.
	IncResourceDeleteEventProcessedError(handler string)
	// ObserveDurationResourceAddEventProcessedSuccess measures the duration it took to process
	// until now a successful processed add event.
	ObserveDurationResourceAddEventProcessedSuccess(handler string, start time.Time)
	// ObserveDurationResourceAddEventProcessedError measures the duration it took to process
	// until now a failed processed add event.
	ObserveDurationResourceAddEventProcessedError(handler string, start time.Time)
	// ObserveDurationResourceAddEventProcessedSuccess measures the duration it took to process
	// until now a successful processed delete event.
	ObserveDurationResourceDeleteEventProcessedSuccess(handler string, start time.Time)
	// ObserveDurationResourceAddEventProcessedError measures the duration it took to process
	// until now a failed processed delete event.
	ObserveDurationResourceDeleteEventProcessedError(handler string, start time.Time)
}
