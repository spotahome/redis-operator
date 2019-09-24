package k8s

import (
	"testing"

	"github.com/spotahome/redis-operator/log"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubernetes "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
)

var (
	secretsGroup = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
)

func newSecretGetAction(ns, name string) kubetesting.GetActionImpl {
	return kubetesting.NewGetAction(secretsGroup, ns, name)
}

func TestSecretServiceGet(t *testing.T) {

	t.Run("Test getting a secret", func(t *testing.T) {
		assert := assert.New(t)

		secret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test_secret",
				Namespace: "test_namespace",
			},
			Data: map[string][]byte{
				"foo": []byte("bar"),
			},
		}

		mcli := &kubernetes.Clientset{}
		mcli.AddReactor("create", "secrets", func(action kubetesting.Action) (bool, runtime.Object, error) {
			return true, &secret, nil
		})
		mcli.AddReactor("get", "secrets", func(action kubetesting.Action) (bool, runtime.Object, error) {
			a := (action).(kubetesting.GetActionImpl)
			if a.Namespace == secret.ObjectMeta.Namespace && a.Name == secret.ObjectMeta.Name {
				return true, &secret, nil
			}
			return true, nil, errors.NewNotFound(action.GetResource().GroupResource(), a.Name)
		})

		_, err := mcli.CoreV1().Secrets(secret.ObjectMeta.Namespace).Create(&secret)
		assert.NoError(err)

		// test getting the secret
		service := NewSecretService(mcli, log.Dummy)
		ss, err := service.GetSecret(secret.ObjectMeta.Namespace, secret.ObjectMeta.Name)
		assert.NotNil(ss)
		assert.NoError(err)

		// test getting a nonexistent secret
		_, err = service.GetSecret(secret.ObjectMeta.Namespace, secret.ObjectMeta.Name+"nonexistent")
		assert.Error(err)
		assert.True(errors.IsNotFound(err))
	})
}
