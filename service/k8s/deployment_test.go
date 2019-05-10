package k8s_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
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
	deploymentsGroup = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
)

func newDeploymentUpdateAction(ns string, deployment *appsv1.Deployment) kubetesting.UpdateActionImpl {
	return kubetesting.NewUpdateAction(deploymentsGroup, ns, deployment)
}

func newDeploymentGetAction(ns, name string) kubetesting.GetActionImpl {
	return kubetesting.NewGetAction(deploymentsGroup, ns, name)
}

func newDeploymentCreateAction(ns string, deployment *appsv1.Deployment) kubetesting.CreateActionImpl {
	return kubetesting.NewCreateAction(deploymentsGroup, ns, deployment)
}

func TestDeploymentServiceGetCreateOrUpdate(t *testing.T) {
	testDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "testdeployment1",
			ResourceVersion: "10",
		},
	}

	testns := "testns"

	tests := []struct {
		name                string
		deployment          *appsv1.Deployment
		getDeploymentResult *appsv1.Deployment
		errorOnGet          error
		errorOnCreation     error
		expActions          []kubetesting.Action
		expErr              bool
	}{
		{
			name:                "A new deployment should create a new deployment.",
			deployment:          testDeployment,
			getDeploymentResult: nil,
			errorOnGet:          kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation:     nil,
			expActions: []kubetesting.Action{
				newDeploymentGetAction(testns, testDeployment.ObjectMeta.Name),
				newDeploymentCreateAction(testns, testDeployment),
			},
			expErr: false,
		},
		{
			name:                "A new deployment should error when create a new deployment fails.",
			deployment:          testDeployment,
			getDeploymentResult: nil,
			errorOnGet:          kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation:     errors.New("wanted error"),
			expActions: []kubetesting.Action{
				newDeploymentGetAction(testns, testDeployment.ObjectMeta.Name),
				newDeploymentCreateAction(testns, testDeployment),
			},
			expErr: true,
		},
		{
			name:                "An existent deployment should update the deployment.",
			deployment:          testDeployment,
			getDeploymentResult: testDeployment,
			errorOnGet:          nil,
			errorOnCreation:     nil,
			expActions: []kubetesting.Action{
				newDeploymentGetAction(testns, testDeployment.ObjectMeta.Name),
				newDeploymentUpdateAction(testns, testDeployment),
			},
			expErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock.
			mcli := &kubernetes.Clientset{}
			mcli.AddReactor("get", "deployments", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, test.getDeploymentResult, test.errorOnGet
			})
			mcli.AddReactor("create", "deployments", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, nil, test.errorOnCreation
			})

			service := k8s.NewDeploymentService(mcli, log.Dummy)
			err := service.CreateOrUpdateDeployment(testns, test.deployment)

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
