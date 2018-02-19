package k8s_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubernetes "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"

	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/service/k8s"
)

var (
	podsGroup = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
)

func newPodUpdateAction(ns string, pod *corev1.Pod) kubetesting.UpdateActionImpl {
	return kubetesting.NewUpdateAction(podsGroup, ns, pod)
}

func newPodGetAction(ns, name string) kubetesting.GetActionImpl {
	return kubetesting.NewGetAction(podsGroup, ns, name)
}

func newPodCreateAction(ns string, pod *corev1.Pod) kubetesting.CreateActionImpl {
	return kubetesting.NewCreateAction(podsGroup, ns, pod)
}

func TestPodServiceGetCreateOrUpdate(t *testing.T) {
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "testpod1",
			ResourceVersion: "10",
		},
	}

	testns := "testns"

	tests := []struct {
		name            string
		pod             *corev1.Pod
		getPodResult    *corev1.Pod
		errorOnGet      error
		errorOnCreation error
		expActions      []kubetesting.Action
		expErr          bool
	}{
		{
			name:            "A new pod should create a new pod.",
			pod:             testPod,
			getPodResult:    nil,
			errorOnGet:      kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation: nil,
			expActions: []kubetesting.Action{
				newPodGetAction(testns, testPod.ObjectMeta.Name),
				newPodCreateAction(testns, testPod),
			},
			expErr: false,
		},
		{
			name:            "A new pod should error when create a new pod fails.",
			pod:             testPod,
			getPodResult:    nil,
			errorOnGet:      kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation: errors.New("wanted error"),
			expActions: []kubetesting.Action{
				newPodGetAction(testns, testPod.ObjectMeta.Name),
				newPodCreateAction(testns, testPod),
			},
			expErr: true,
		},
		{
			name:            "An existent pod should update the pod.",
			pod:             testPod,
			getPodResult:    testPod,
			errorOnGet:      nil,
			errorOnCreation: nil,
			expActions: []kubetesting.Action{
				newPodGetAction(testns, testPod.ObjectMeta.Name),
				newPodUpdateAction(testns, testPod),
			},
			expErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock.
			mcli := &kubernetes.Clientset{}
			mcli.AddReactor("get", "pods", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, test.getPodResult, test.errorOnGet
			})
			mcli.AddReactor("create", "pods", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, nil, test.errorOnCreation
			})

			service := k8s.NewPodService(mcli, log.Dummy)
			err := service.CreateOrUpdatePod(testns, test.pod)

			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				// Check calls to kubernetes.
				assert.Equal(test.expActions, mcli.Actions())
			}
		})
	}
}
