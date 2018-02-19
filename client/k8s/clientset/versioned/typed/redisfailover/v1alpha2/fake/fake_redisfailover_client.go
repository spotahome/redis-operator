package fake

import (
	v1alpha2 "github.com/spotahome/redis-operator/client/k8s/clientset/versioned/typed/redisfailover/v1alpha2"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeStorageV1alpha2 struct {
	*testing.Fake
}

func (c *FakeStorageV1alpha2) RedisFailovers(namespace string) v1alpha2.RedisFailoverInterface {
	return &FakeRedisFailovers{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeStorageV1alpha2) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
