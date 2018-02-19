package fake

import (
	v1alpha1 "github.com/spotahome/kooper/examples/pod-terminator-operator/client/k8s/clientset/versioned/typed/chaos/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeChaosV1alpha1 struct {
	*testing.Fake
}

func (c *FakeChaosV1alpha1) PodTerminators() v1alpha1.PodTerminatorInterface {
	return &FakePodTerminators{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeChaosV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
