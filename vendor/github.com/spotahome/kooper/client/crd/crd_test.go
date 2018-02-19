package crd_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionscli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	kubetesting "k8s.io/client-go/testing"

	"github.com/spotahome/kooper/client/crd"
	"github.com/spotahome/kooper/log"
	mtime "github.com/spotahome/kooper/mocks/wrapper/time"
)

var (
	crdGroup           = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1beta1", Resource: "customresourcedefinitions"}
	goodClusterVersion = "v1.7"
	badClusterVersion  = "v1.6"
)

// newClient returns a new mock client.
func newClient() *apiextensionscli.Clientset {
	return newVersionedClusterClient(goodClusterVersion)
}

// newVersionedClusterClient returns a new mock client with the cluster version set
func newVersionedClusterClient(clusterVersion string) *apiextensionscli.Clientset {
	cli := apiextensionscli.NewSimpleClientset()

	// Fake cluster version.
	fakeDiscovery, _ := cli.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = &version.Info{
		GitVersion: clusterVersion,
	}

	// Use different fake action registry for CRD api calls client and discovery client. By default they
	// share the same registry(testing.Fake) and we don't want to mix version check actions with the CRD
	// actions.
	cli.Fake = kubetesting.Fake{}

	return cli
}

func newCRDGetAction(name string) kubetesting.GetActionImpl {
	return kubetesting.NewGetAction(crdGroup, "", name)
}

func newCRDCreateAction(crd *apiextensionsv1beta1.CustomResourceDefinition) kubetesting.CreateActionImpl {
	return kubetesting.NewCreateAction(crdGroup, "", crd)
}

func TestCRDEnsurePresent(t *testing.T) {
	tests := []struct {
		name           string
		clusterVersion string
		crd            crd.Conf
		retErr         error
		expErr         bool
		expCalls       []kubetesting.Action
	}{
		{
			name:           "Creating a non existen CRD should create a crd without error",
			clusterVersion: goodClusterVersion,
			crd: crd.Conf{
				Kind:       "Test",
				NamePlural: "tests",
				Scope:      crd.ClusterScoped,
				Group:      "toilettesting",
				Version:    "v99",
			},
			retErr: nil,
			expErr: false,
			expCalls: []kubetesting.Action{
				newCRDCreateAction(&apiextensionsv1beta1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tests.toilettesting",
					},
					Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
						Group:   "toilettesting",
						Version: "v99",
						Scope:   crd.ClusterScoped,
						Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
							Plural: "tests",
							Kind:   "Test",
						},
					},
				}),
				newCRDGetAction("tests.toilettesting"),
			},
		},
		{
			name:           "Creating a CRD that errors should return an error.",
			clusterVersion: goodClusterVersion,
			crd: crd.Conf{
				Kind:       "Test",
				NamePlural: "tests",
				Scope:      crd.ClusterScoped,
				Group:      "toilettesting",
				Version:    "v99",
			},
			retErr:   errors.New("wanted error"),
			expErr:   true,
			expCalls: []kubetesting.Action{},
		},
		{
			name:           "Creating a CRD that exists shouldn't return an error.",
			clusterVersion: goodClusterVersion,
			crd: crd.Conf{
				Kind:       "Test",
				NamePlural: "tests",
				Scope:      crd.ClusterScoped,
				Group:      "toilettesting",
				Version:    "v99",
			},
			expCalls: []kubetesting.Action{
				newCRDCreateAction(&apiextensionsv1beta1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tests.toilettesting",
					},
					Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
						Group:   "toilettesting",
						Version: "v99",
						Scope:   crd.ClusterScoped,
						Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
							Plural: "tests",
							Kind:   "Test",
						},
					},
				}),
			},
			retErr: kubeerrors.NewAlreadyExists(schema.GroupResource{}, ""),
			expErr: false,
		},
		{
			name:           "If the cluster version is not supported it should fail.",
			clusterVersion: badClusterVersion,
			crd:            crd.Conf{},
			retErr:         nil,
			expErr:         true,
			expCalls:       []kubetesting.Action{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock calls
			cli := newVersionedClusterClient(test.clusterVersion)
			cli.AddReactor("create", "customresourcedefinitions", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, nil, test.retErr
			})
			cli.AddReactor("get", "customresourcedefinitions", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			crdCli := crd.NewClient(cli, log.Dummy)
			err := crdCli.EnsurePresent(test.crd)
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expCalls, cli.Actions())
			}

		})
	}
}

func TestCRDWaitToBePresent(t *testing.T) {
	never := 999999 * time.Hour
	tests := []struct {
		name           string
		clusterVersion string
		crdName        string
		timeout        time.Duration
		interval       time.Duration
		readyAfter     time.Duration
		expErr         bool
	}{
		{
			name:           "If timeouts it should return an error",
			clusterVersion: goodClusterVersion,
			crdName:        "test",
			timeout:        1,
			interval:       never,
			readyAfter:     never,
			expErr:         true,
		},
		{
			name:           "If the cluster version is not supported, it should fail",
			clusterVersion: badClusterVersion,
			crdName:        "test",
			timeout:        1,
			interval:       never,
			readyAfter:     never,
			expErr:         true,
		},
		{
			name:           "After some time the CRD will be ready and won't timeout",
			clusterVersion: goodClusterVersion,
			crdName:        "test",
			timeout:        50 * time.Millisecond,
			interval:       5 * time.Millisecond,
			readyAfter:     25 * time.Millisecond,
			expErr:         false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock calls
			mt := &mtime.Time{}
			mt.On("After", mock.Anything).Once().Return(time.After(test.timeout))
			mt.On("NewTicker", mock.Anything).Once().Return(time.NewTicker(test.interval))
			cli := newVersionedClusterClient(test.clusterVersion)
			start := time.Now()
			cli.AddReactor("get", "customresourcedefinitions", func(action kubetesting.Action) (bool, runtime.Object, error) {
				// Get how long we've been running and if it passed readyAfter
				// our CRD is ready.
				runningTime := time.Now().Sub(start)
				if runningTime >= test.readyAfter {
					return true, nil, nil
				}
				return true, nil, fmt.Errorf("wanted error")

			})

			crdCli := crd.NewCustomClient(cli, mt, log.Dummy)
			err := crdCli.WaitToBePresent(test.crdName, 0)
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
