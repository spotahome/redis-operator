package metrics

import "time"

// Dummy is a dummy stats recorder.
var Dummy = &dummy{}

type dummy struct{}

func (d *dummy) IncResourceDeleteEventQueued(_ string)                                    {}
func (d *dummy) IncResourceAddEventQueued(_ string)                                       {}
func (d *dummy) IncResourceAddEventProcessedSuccess(_ string)                             {}
func (d *dummy) IncResourceAddEventProcessedError(_ string)                               {}
func (d *dummy) IncResourceDeleteEventProcessedSuccess(_ string)                          {}
func (d *dummy) IncResourceDeleteEventProcessedError(_ string)                            {}
func (d *dummy) ObserveDurationResourceAddEventProcessedSuccess(_ string, _ time.Time)    {}
func (d *dummy) ObserveDurationResourceAddEventProcessedError(_ string, _ time.Time)      {}
func (d *dummy) ObserveDurationResourceDeleteEventProcessedSuccess(_ string, _ time.Time) {}
func (d *dummy) ObserveDurationResourceDeleteEventProcessedError(_ string, _ time.Time)   {}
