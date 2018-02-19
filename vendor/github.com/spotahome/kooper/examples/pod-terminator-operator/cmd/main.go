package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spotahome/kooper/client/crd"
	applogger "github.com/spotahome/kooper/log"
	apiextensionscli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	podtermk8scli "github.com/spotahome/kooper/examples/pod-terminator-operator/client/k8s/clientset/versioned"
	"github.com/spotahome/kooper/examples/pod-terminator-operator/log"
	"github.com/spotahome/kooper/examples/pod-terminator-operator/operator"
)

// Main is the main program.
type Main struct {
	flags  *Flags
	config operator.Config
	logger log.Logger
}

// New returns the main application.
func New(logger log.Logger) *Main {
	f := NewFlags()
	return &Main{
		flags:  f,
		config: f.OperatorConfig(),
		logger: logger,
	}
}

// Run runs the app.
func (m *Main) Run(stopC <-chan struct{}) error {
	m.logger.Infof("initializing pod termination operator")

	// Get kubernetes rest client.
	ptCli, crdCli, k8sCli, err := m.getKubernetesClients()
	if err != nil {
		return err
	}

	// Create the operator and run
	op, err := operator.New(m.config, ptCli, crdCli, k8sCli, m.logger)
	if err != nil {
		return err
	}

	return op.Run(stopC)
}

// getKubernetesClients returns all the required clients to communicate with
// kubernetes cluster: CRD type client, pod terminator types client, kubernetes core types client.
func (m *Main) getKubernetesClients() (podtermk8scli.Interface, crd.Interface, kubernetes.Interface, error) {
	var err error
	var cfg *rest.Config

	// If devel mode then use configuration flag path.
	if m.flags.Development {
		cfg, err = clientcmd.BuildConfigFromFlags("", m.flags.KubeConfig)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("could not load configuration: %s", err)
		}
	} else {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error loading kubernetes configuration inside cluster, check app is running outside kubernetes cluster or run in development mode: %s", err)
		}
	}

	// Create clients.
	k8sCli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	// App CRD k8s types client.
	ptCli, err := podtermk8scli.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	// CRD cli.
	aexCli, err := apiextensionscli.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, err
	}
	crdCli := crd.NewClient(aexCli, m.logger)

	return ptCli, crdCli, k8sCli, nil
}

func main() {
	logger := &applogger.Std{}

	stopC := make(chan struct{})
	finishC := make(chan error)
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGTERM, syscall.SIGINT)
	m := New(logger)

	// Run in background the operator.
	go func() {
		finishC <- m.Run(stopC)
	}()

	select {
	case err := <-finishC:
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running operator: %s", err)
			os.Exit(1)
		}
	case <-signalC:
		logger.Infof("Signal captured, exiting...")
	}

}
