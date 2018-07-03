// +build integration

package controller_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/operator/controller"
	"github.com/spotahome/kooper/operator/handler"
	"github.com/spotahome/kooper/operator/retrieve"
	"github.com/spotahome/kooper/test/integration/helper/cli"
	"github.com/spotahome/kooper/test/integration/helper/prepare"
)

// TestControllerHandleEvents will test the controller receives the resources list and watch
// events are received and handled correctly.
func TestControllerHandleEvents(t *testing.T) {
	tests := []struct {
		name               string
		addServices        []*corev1.Service
		updateServices     []string
		delServices        []string
		expAddedServices   []string
		expDeletedServices []string
	}{
		{
			name: "If a controller is watching services it should react to the service change events.",
			addServices: []*corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "svc1"},
					Spec: corev1.ServiceSpec{
						Type: "ClusterIP",
						Ports: []corev1.ServicePort{
							corev1.ServicePort{Name: "port1", Port: 8080},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "svc2"},
					Spec: corev1.ServiceSpec{
						Type: "ClusterIP",
						Ports: []corev1.ServicePort{
							corev1.ServicePort{Name: "port1", Port: 8080},
						},
					},
				},
			},
			updateServices:     []string{"svc1"},
			delServices:        []string{"svc1", "svc2"},
			expAddedServices:   []string{"svc1", "svc2", "svc1"},
			expDeletedServices: []string{"svc1", "svc2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)
			resync := 30 * time.Second
			stopC := make(chan struct{})
			var gotAddedServices []string
			var gotDeletedServices []string

			// Create the kubernetes client.
			k8scli, _, _, err := cli.GetK8sClients("")

			require.NoError(err, "kubernetes client is required")

			// Prepare the environment on the cluster.
			prep := prepare.New(k8scli, t)
			prep.SetUp()
			defer prep.TearDown()

			// Create the reitrever.
			rt := &retrieve.Resource{
				ListerWatcher: cache.NewListWatchFromClient(k8scli.CoreV1().RESTClient(), "services", prep.Namespace().Name, fields.Everything()),
				Object:        &corev1.Service{},
			}

			// Call times are the number of times the handler should be called before sending the termination signal.
			stopCallTimes := len(test.addServices) + len(test.updateServices) + len(test.delServices)
			calledTimes := 0
			var mx sync.Mutex

			// Create the handler.
			hl := &handler.HandlerFunc{
				AddFunc: func(_ context.Context, obj runtime.Object) error {
					mx.Lock()
					calledTimes++
					mx.Unlock()

					svc := obj.(*corev1.Service)
					gotAddedServices = append(gotAddedServices, svc.Name)
					if calledTimes >= stopCallTimes {
						close(stopC)
					}
					return nil
				},
				DeleteFunc: func(_ context.Context, id string) error {
					mx.Lock()
					calledTimes++
					mx.Unlock()

					// Ignore namespace.
					id = strings.Split(id, "/")[1]
					gotDeletedServices = append(gotDeletedServices, id)
					if calledTimes >= stopCallTimes {
						close(stopC)
					}
					return nil
				},
			}

			// Create a Pod controller.
			ctrl := controller.NewSequential(resync, hl, rt, nil, log.Dummy)
			require.NotNil(ctrl, "controller is required")
			go ctrl.Run(stopC)

			// Create the required services.
			for _, svc := range test.addServices {
				_, err := k8scli.CoreV1().Services(prep.Namespace().Name).Create(svc)
				assert.NoError(err)
				time.Sleep(1 * time.Second)
			}

			for _, svc := range test.updateServices {
				origSvc, err := k8scli.CoreV1().Services(prep.Namespace().Name).Get(svc, metav1.GetOptions{})
				if assert.NoError(err) {
					// Change something
					origSvc.Spec.Ports = append(origSvc.Spec.Ports, corev1.ServicePort{Name: "updateport", Port: 9876})
					_, err := k8scli.CoreV1().Services(prep.Namespace().Name).Update(origSvc)
					assert.NoError(err)
					time.Sleep(1 * time.Second)
				}
			}

			// Delete the required services.
			for _, svc := range test.delServices {
				err := k8scli.CoreV1().Services(prep.Namespace().Name).Delete(svc, &metav1.DeleteOptions{})
				assert.NoError(err)
				time.Sleep(1 * time.Second)
			}

			// Wait until we have finished.
			select {
			// Timeout.
			case <-time.After(20 * time.Second):
			// Finished.
			case <-stopC:
			}

			// Check.
			assert.Equal(test.expAddedServices, gotAddedServices)
			assert.Equal(test.expDeletedServices, gotDeletedServices)
		})
	}
}
