package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	applogger "github.com/spotahome/kooper/log"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spotahome/kooper/examples/echo-pod-controller/controller"
	"github.com/spotahome/kooper/examples/echo-pod-controller/log"
)

// Main is the main program.
type Main struct {
	flags  *Flags
	config controller.Config
	logger log.Logger
}

// New returns the main application.
func New(logger log.Logger) *Main {
	f := NewFlags()
	return &Main{
		flags:  f,
		config: f.ControllerConfig(),
		logger: logger,
	}
}

// Run runs the app.
func (m *Main) Run(stopC <-chan struct{}) error {
	m.logger.Infof("initializing echo controller")

	// Get kubernetes rest client.
	k8sCli, err := m.getKubernetesClient()
	if err != nil {
		return err
	}

	// Create the controller and run
	ctrl, err := controller.New(m.config, k8sCli, m.logger)
	if err != nil {
		return err
	}

	return ctrl.Run(stopC)
}

func (m *Main) getKubernetesClient() (kubernetes.Interface, error) {
	var err error
	var cfg *rest.Config

	// If devel mode then use configuration flag path.
	if m.flags.Development {
		cfg, err = clientcmd.BuildConfigFromFlags("", m.flags.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("could not load configuration: %s", err)
		}
	} else {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("error loading kubernetes configuration inside cluster, check app is running outside kubernetes cluster or run in development mode: %s", err)
		}
	}

	return kubernetes.NewForConfig(cfg)
}

func main() {
	logger := &applogger.Std{}

	stopC := make(chan struct{})
	finishC := make(chan error)
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGTERM, syscall.SIGINT)
	m := New(logger)

	// Run in background the controller.
	go func() {
		finishC <- m.Run(stopC)
	}()

	select {
	case err := <-finishC:
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running controller: %s", err)
			os.Exit(1)
		}
	case <-signalC:
		logger.Infof("Signal captured, exiting...")
	}

}
