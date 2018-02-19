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
	configMapsGroup = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
)

func newConfigMapUpdateAction(ns string, configMap *corev1.ConfigMap) kubetesting.UpdateActionImpl {
	return kubetesting.NewUpdateAction(configMapsGroup, ns, configMap)
}

func newConfigMapGetAction(ns, name string) kubetesting.GetActionImpl {
	return kubetesting.NewGetAction(configMapsGroup, ns, name)
}

func newConfigMapCreateAction(ns string, configMap *corev1.ConfigMap) kubetesting.CreateActionImpl {
	return kubetesting.NewCreateAction(configMapsGroup, ns, configMap)
}

func TestConfigMapServiceGetCreateOrUpdate(t *testing.T) {
	testConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "testconfigmap1",
			ResourceVersion: "10",
		},
	}

	testns := "testns"

	tests := []struct {
		name               string
		configMap          *corev1.ConfigMap
		getConfigMapResult *corev1.ConfigMap
		errorOnGet         error
		errorOnCreation    error
		expActions         []kubetesting.Action
		expErr             bool
	}{
		{
			name:               "A new configmap should create a new configmap.",
			configMap:          testConfigMap,
			getConfigMapResult: nil,
			errorOnGet:         kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation:    nil,
			expActions: []kubetesting.Action{
				newConfigMapGetAction(testns, testConfigMap.ObjectMeta.Name),
				newConfigMapCreateAction(testns, testConfigMap),
			},
			expErr: false,
		},
		{
			name:               "A new configmap should error when create a new configmap fails.",
			configMap:          testConfigMap,
			getConfigMapResult: nil,
			errorOnGet:         kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation:    errors.New("wanted error"),
			expActions: []kubetesting.Action{
				newConfigMapGetAction(testns, testConfigMap.ObjectMeta.Name),
				newConfigMapCreateAction(testns, testConfigMap),
			},
			expErr: true,
		},
		{
			name:               "An existent configmap should update the configmap.",
			configMap:          testConfigMap,
			getConfigMapResult: testConfigMap,
			errorOnGet:         nil,
			errorOnCreation:    nil,
			expActions: []kubetesting.Action{
				newConfigMapGetAction(testns, testConfigMap.ObjectMeta.Name),
				newConfigMapUpdateAction(testns, testConfigMap),
			},
			expErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock.
			mcli := &kubernetes.Clientset{}
			mcli.AddReactor("get", "configmaps", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, test.getConfigMapResult, test.errorOnGet
			})
			mcli.AddReactor("create", "configmaps", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, nil, test.errorOnCreation
			})

			service := k8s.NewConfigMapService(mcli, log.Dummy)
			err := service.CreateOrUpdateConfigMap(testns, test.configMap)

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
