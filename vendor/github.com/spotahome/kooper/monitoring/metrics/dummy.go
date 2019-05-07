package metrics

import "time"

// Dummy is a dummy stats recorder.
var Dummy = &dummy{}

type dummy struct{}

func (*dummy) IncResourceEventQueued(_ string, _ EventType) {
}
func (*dummy) IncResourceEventProcessed(_ string, _ EventType) {
}
func (*dummy) IncResourceEventProcessedError(_ string, _ EventType) {
}
func (*dummy) ObserveDurationResourceEventProcessed(_ string, _ EventType, _ time.Time) {
}
