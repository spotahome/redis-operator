package metrics

import (
	koopercontroller "github.com/spotahome/kooper/v2/controller"
)

// Dummy is a handy instnce of a dummy instrumenter, most of the times it will be used on tests.
var Dummy = &dummy{
	MetricsRecorder: koopercontroller.DummyMetricsRecorder,
}

// dummy is a dummy implementation of Instrumenter.
type dummy struct {
	koopercontroller.MetricsRecorder
}

func (d *dummy) SetClusterOK(namespace string, name string)               {}
func (d *dummy) SetClusterError(namespace string, name string)            {}
func (d *dummy) DeleteCluster(namespace string, name string)              {}
func (d *dummy) SetRedisInstance(IP string, masterIP string, role string) {}
func (d *dummy) ResetRedisInstance()                                      {}
func (d *dummy) RecordEnsureOperation(objectNamespace string, objectName string, objectKind string, resourceName string, status string) {
}
func (d *dummy) RecordRedisCheck(namespace string, resource string, indicator string, instance string, status string) {
}
func (d *dummy) RecordSentinelCheck(namespace string, resource string, indicator string, instance string, status string) {
}
func (d dummy) RecordK8sOperation(namespace string, kind string, object string, operation string, status string, err string) {
}
func (d dummy) RecordRedisOperation(kind string, IP string, operation string, status string, err string) {
}
