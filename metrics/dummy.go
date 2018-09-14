package metrics

// Dummy is a handy instnce of a dummy instrumenter, most of the times it will be used on tests.
var Dummy = &dummy{}

// dummy is a dummy implementation of Instrumenter.
type dummy struct{}

func (d *dummy) SetClusterOK(namespace string, name string)    {}
func (d *dummy) SetClusterError(namespace string, name string) {}
func (d *dummy) DeleteCluster(namespace string, name string)   {}
