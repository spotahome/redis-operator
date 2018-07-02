package prepare

import (
	"strings"
	"testing"

	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Preparer knows ho to prepare a test case in a kubernetes cluster.
type Preparer struct {
	t         *testing.T
	cli       kubernetes.Interface
	namespace *corev1.Namespace
}

// New returns a new preparer.
func New(cli kubernetes.Interface, t *testing.T) *Preparer {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(randomdata.SillyName()),
		},
	}
	return &Preparer{
		t:         t,
		cli:       cli,
		namespace: ns,
	}
}

// SetUp will set up all the required preparation for the tests in a Kubernetes cluster.
func (p *Preparer) SetUp() {
	ns, err := p.cli.CoreV1().Namespaces().Create(p.namespace)
	require.NoError(p.t, err, "set up failed, can't continue without setting up an environment")
	p.namespace = ns
}

// Namespace returns the namespace where the environment is set.
func (p *Preparer) Namespace() *corev1.Namespace {
	return p.namespace
}

// TearDown will tear down all the required preparation for the tests in a Kubernetes cluster.
func (p *Preparer) TearDown() {
	err := p.cli.CoreV1().Namespaces().Delete(p.namespace.Name, &metav1.DeleteOptions{})
	require.NoError(p.t, err, "tear down failed, can't continue without destroying an environment")
}
