package k8s_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
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
	rbGroup = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}
)

func newRBUpdateAction(ns string, rb *rbacv1.RoleBinding) kubetesting.UpdateActionImpl {
	return kubetesting.NewUpdateAction(rbGroup, ns, rb)
}

func newRBGetAction(ns, name string) kubetesting.GetActionImpl {
	return kubetesting.NewGetAction(rbGroup, ns, name)
}

func newRBCreateAction(ns string, rb *rbacv1.RoleBinding) kubetesting.CreateActionImpl {
	return kubetesting.NewCreateAction(rbGroup, ns, rb)
}
func newRBDeleteAction(ns string, name string) kubetesting.DeleteActionImpl {
	return kubetesting.NewDeleteAction(rbGroup, ns, name)
}

func TestRBACServiceGetCreateOrUpdateRoleBinding(t *testing.T) {
	testRB := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test1",
			ResourceVersion: "15",
		},
		RoleRef: rbacv1.RoleRef{
			Name: "test1",
		},
	}

	testns := "testns"

	tests := []struct {
		name            string
		rb              *rbacv1.RoleBinding
		getRBResult     *rbacv1.RoleBinding
		errorOnGet      error
		errorOnCreation error
		expActions      []kubetesting.Action
		expErr          bool
	}{
		{
			name:            "A new role binding should create a new role binding.",
			rb:              testRB,
			getRBResult:     nil,
			errorOnGet:      kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation: nil,
			expActions: []kubetesting.Action{
				newRBGetAction(testns, testRB.ObjectMeta.Name),
				newRBCreateAction(testns, testRB),
			},
			expErr: false,
		},
		{
			name:            "A new role binding should error when create a new role binding fails.",
			rb:              testRB,
			getRBResult:     nil,
			errorOnGet:      kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation: errors.New("wanted error"),
			expActions: []kubetesting.Action{
				newRBGetAction(testns, testRB.ObjectMeta.Name),
				newRBUpdateAction(testns, testRB),
			},
			expErr: true,
		},
		{
			name:            "An existent role binding should update the role binding.",
			rb:              testRB,
			getRBResult:     testRB,
			errorOnGet:      nil,
			errorOnCreation: nil,
			expActions: []kubetesting.Action{
				newRBGetAction(testns, testRB.ObjectMeta.Name),
				newRBUpdateAction(testns, testRB),
			},
			expErr: false,
		},
		{
			name: "An change in role reference inside binding should recreate the role binding.",
			rb:   testRB,
			getRBResult: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test1",
					ResourceVersion: "15",
				},
				RoleRef: rbacv1.RoleRef{
					Name: "oldroleRef",
				},
			},
			errorOnGet:      nil,
			errorOnCreation: nil,
			expActions: []kubetesting.Action{
				newRBGetAction(testns, testRB.ObjectMeta.Name),
				newRBDeleteAction(testns, testRB.Name),
				newRBCreateAction(testns, testRB),
			},
			expErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock.
			mcli := &kubernetes.Clientset{}
			mcli.AddReactor("get", "rolebindings", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, test.getRBResult, test.errorOnGet
			})
			mcli.AddReactor("create", "rolebindings", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, nil, test.errorOnCreation
			})

			service := k8s.NewRBACService(mcli, log.Dummy)
			err := service.CreateOrUpdateRoleBinding(testns, test.rb)

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
