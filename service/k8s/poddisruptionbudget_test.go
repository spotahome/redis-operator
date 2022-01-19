package k8s_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
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
	podDisruptionBudgetsGroup = schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "poddisruptionbudgets"}
)

func newPodDisruptionBudgetUpdateAction(ns string, podDisruptionBudget *policyv1beta1.PodDisruptionBudget) kubetesting.UpdateActionImpl {
	return kubetesting.NewUpdateAction(podDisruptionBudgetsGroup, ns, podDisruptionBudget)
}

func newPodDisruptionBudgetGetAction(ns, name string) kubetesting.GetActionImpl {
	return kubetesting.NewGetAction(podDisruptionBudgetsGroup, ns, name)
}

func newPodDisruptionBudgetCreateAction(ns string, podDisruptionBudget *policyv1beta1.PodDisruptionBudget) kubetesting.CreateActionImpl {
	return kubetesting.NewCreateAction(podDisruptionBudgetsGroup, ns, podDisruptionBudget)
}

func TestPodDisruptionBudgetServiceGetCreateOrUpdate(t *testing.T) {
	testPodDisruptionBudget := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "testpodDisruptionBudget1",
			ResourceVersion: "10",
		},
	}

	testns := "testns"

	tests := []struct {
		name                         string
		podDisruptionBudget          *policyv1beta1.PodDisruptionBudget
		getPodDisruptionBudgetResult *policyv1beta1.PodDisruptionBudget
		errorOnGet                   error
		errorOnCreation              error
		expActions                   []kubetesting.Action
		expErr                       bool
	}{
		{
			name:                         "A new podDisruptionBudget should create a new podDisruptionBudget.",
			podDisruptionBudget:          testPodDisruptionBudget,
			getPodDisruptionBudgetResult: nil,
			errorOnGet:                   kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation:              nil,
			expActions: []kubetesting.Action{
				newPodDisruptionBudgetGetAction(testns, testPodDisruptionBudget.ObjectMeta.Name),
				newPodDisruptionBudgetCreateAction(testns, testPodDisruptionBudget),
			},
			expErr: false,
		},
		{
			name:                         "A new podDisruptionBudget should error when create a new podDisruptionBudget fails.",
			podDisruptionBudget:          testPodDisruptionBudget,
			getPodDisruptionBudgetResult: nil,
			errorOnGet:                   kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation:              errors.New("wanted error"),
			expActions: []kubetesting.Action{
				newPodDisruptionBudgetGetAction(testns, testPodDisruptionBudget.ObjectMeta.Name),
				newPodDisruptionBudgetCreateAction(testns, testPodDisruptionBudget),
			},
			expErr: true,
		},
		{
			name:                         "An existent podDisruptionBudget should update the podDisruptionBudget.",
			podDisruptionBudget:          testPodDisruptionBudget,
			getPodDisruptionBudgetResult: testPodDisruptionBudget,
			errorOnGet:                   nil,
			errorOnCreation:              nil,
			expActions: []kubetesting.Action{
				newPodDisruptionBudgetGetAction(testns, testPodDisruptionBudget.ObjectMeta.Name),
				newPodDisruptionBudgetUpdateAction(testns, testPodDisruptionBudget),
			},
			expErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock.
			mcli := &kubernetes.Clientset{}
			mcli.AddReactor("get", "poddisruptionbudgets", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, test.getPodDisruptionBudgetResult, test.errorOnGet
			})
			mcli.AddReactor("create", "poddisruptionbudgets", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, nil, test.errorOnCreation
			})

			service := k8s.NewPodDisruptionBudgetService(mcli, log.Dummy)
			err := service.CreateOrUpdatePodDisruptionBudget(testns, test.podDisruptionBudget)

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
