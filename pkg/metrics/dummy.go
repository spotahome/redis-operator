package metrics

// Dummy is a handy instnce of a dummy instrumenter, most of the times it will be used on tests.
var Dummy = &dummy{}

// dummy is a dummy implementation of Instrumenter.
type dummy struct{}

func (d *dummy) SetClustersRunning(n float64)          {}
func (d *dummy) SetClustersCreating(n float64)         {}
func (d *dummy) SetClustersFailed(n float64)           {}
func (d *dummy) IncAddEventHandled(failover string)    {}
func (d *dummy) IncUpdateEventHandled(failover string) {}
func (d *dummy) IncDeleteEventHandled(failover string) {}
func (d *dummy) SetClusterMasters(n float64, cluster string){}
