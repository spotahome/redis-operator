package options

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"

	v1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	fakeclientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned/fake"
	"github.com/spotahome/redis-operator/pkg/kubectl-redis-failover/options"
)

const (
	apiVersion = "databases.spotahome.com/v1"
)

// NewFakeRedisFailoverOptions returns a options.ArgoRolloutsOptions suitable for testing
func NewFakeRedisFailoverOptions(obj ...runtime.Object) (*cmdtesting.TestFactory, *options.RedisFailoverOptions) {
	iostreams, _, _, _ := genericclioptions.NewTestIOStreams()
	tf := cmdtesting.NewTestFactory()
	o := options.NewRedisFailoverOptions(iostreams)
	o.RESTClientGetter = tf

	var redisFailoverObjs []runtime.Object
	var kubeObjs []runtime.Object
	var allObjs []runtime.Object

	// Loop through supplied fake objects. Set TypeMeta if it wasn't set in the test
	// so that the objects can also go into the fake dynamic client
	for _, o := range obj {
		switch typedO := o.(type) {
		case *v1.RedisFailover:
			typedO.TypeMeta = metav1.TypeMeta{
				Kind:       v1.RFKind,
				APIVersion: apiVersion,
			}
			redisFailoverObjs = append(redisFailoverObjs, o)
		default:
			kubeObjs = append(kubeObjs, o)
		}
		allObjs = append(allObjs, o)
	}

	o.RedisFailoverClient = fakeclientset.NewSimpleClientset(redisFailoverObjs...)
	o.KubeClient = k8sfake.NewSimpleClientset(kubeObjs...)
	err := v1.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}
	listMapping := map[schema.GroupVersionResource]string{
		v1.RedisFailoverGVR: v1.RFKind + "List",
	}

	o.DynamicClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme.Scheme, listMapping, allObjs...)
	return tf, o
}
