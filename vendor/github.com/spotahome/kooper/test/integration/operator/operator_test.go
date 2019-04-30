// +build integration

package operator_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/spotahome/kooper/client/crd"
	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/operator"
	"github.com/spotahome/kooper/operator/controller"
	"github.com/spotahome/kooper/operator/handler"
	"github.com/spotahome/kooper/test/integration/helper/cli"
	"github.com/spotahome/kooper/test/integration/helper/prepare"
	superherov1alpha1 "github.com/spotahome/kooper/test/integration/operator/apis/superhero/v1alpha1"
	integrationtestk8scli "github.com/spotahome/kooper/test/integration/operator/client/k8s/clientset/versioned"
	apiextensionscli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

type spidermanCRD struct {
	aexcli                apiextensionscli.Interface
	crdcli                crd.Interface
	kubeccli              kubernetes.Interface
	integrationtestk8scli integrationtestk8scli.Interface
}

func (s *spidermanCRD) GetListerWatcher() cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return s.integrationtestk8scli.SuperheroV1alpha1().Spidermans("").List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return s.integrationtestk8scli.SuperheroV1alpha1().Spidermans("").Watch(options)
		},
	}
}

func (s *spidermanCRD) GetObject() runtime.Object {
	return &superherov1alpha1.Spiderman{}
}

// podTerminatorCRD satisfies resource.crd interface.
func (s *spidermanCRD) Initialize() error {
	crd := crd.Conf{
		Kind:       superherov1alpha1.SpidermanKind,
		NamePlural: superherov1alpha1.SpidermanNamePlural,
		ShortNames: superherov1alpha1.SpidermanShortNames,
		Group:      superherov1alpha1.SchemeGroupVersion.Group,
		Version:    superherov1alpha1.SchemeGroupVersion.Version,
		Scope:      superherov1alpha1.SpidermanScope,
	}

	return s.crdcli.EnsurePresent(crd)
}

func (s *spidermanCRD) deleteCRD() error {
	crdName := fmt.Sprintf("%s.%s", superherov1alpha1.SpidermanNamePlural, superherov1alpha1.SchemeGroupVersion.Group)
	return s.aexcli.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crdName, &metav1.DeleteOptions{})
}

// TestCRDRegister will test the CRD is registered on the cluster.
func TestCRDRegister(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Starting the operator should register the CRD.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			resync := 30 * time.Second
			stopC := make(chan struct{})

			// Create the kubernetes client.
			k8scli, aexcli, itcli, err := cli.GetK8sClients("")
			require.NoError(err, "kubernetes client is required")

			// Prepare the environment on the cluster.
			prep := prepare.New(k8scli, t)
			prep.SetUp()
			defer prep.TearDown()

			// Create the CRD.
			spcrd := &spidermanCRD{
				aexcli:                aexcli,
				crdcli:                crd.NewClient(aexcli, log.Dummy),
				kubeccli:              k8scli,
				integrationtestk8scli: itcli,
			}

			// Create the handler.
			hl := &handler.HandlerFunc{
				AddFunc: func(_ context.Context, obj runtime.Object) error {
					return nil
				},
				DeleteFunc: func(_ context.Context, id string) error {
					return nil
				},
			}

			// Create a controller.
			ctrl := controller.NewSequential(resync, hl, spcrd, nil, log.Dummy)
			require.NotNil(ctrl, "controller is required")

			// Check no CRD.
			_, err = itcli.Discovery().ServerResourcesForGroupVersion(superherov1alpha1.SchemeGroupVersion.String())
			require.Error(err, "the resource shouldn't be registered")
			// At the end of the test the resource shouldn't be there.
			defer spcrd.deleteCRD()

			// Starting the operator should register the CRD.
			op := operator.NewOperator(spcrd, ctrl, log.Dummy)
			go op.Run(stopC)
			// Stop operator when the test is done.
			defer func() {
				close(stopC)
			}()

			// Wait some time until the registration takes effect.
			<-time.After(2 * time.Second)

			// Check.
			rl, err := itcli.Discovery().ServerResourcesForGroupVersion(superherov1alpha1.SchemeGroupVersion.String())
			if assert.NoError(err) {
				// Check the only resource available is spiderman.
				rr := rl.APIResources
				assert.Len(rr, 1)
				assert.Equal(superherov1alpha1.SpidermanKind, rr[0].Kind)
			}
		})
	}
}
