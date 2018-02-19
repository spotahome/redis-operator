package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	redisfailoverclientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	"github.com/spotahome/redis-operator/operator/redisfailover"
	"github.com/spotahome/redis-operator/service/k8s"
	"github.com/spotahome/redis-operator/service/redis"
)

const (
	gracePeriod = 5 * time.Second
)

// TODO: improve flags.
type cmdFlags struct {
	kubeConfig  string
	development bool
	debug       bool
	listenAddr  string
	metricsPath string
}

func (c *cmdFlags) init() {
	kubehome := filepath.Join(homedir.HomeDir(), ".kube", "config")
	// register flags
	flag.StringVar(&c.kubeConfig, "kubeconfig", kubehome, "kubernetes configuration path, only used when development mode enabled")
	flag.BoolVar(&c.development, "development", false, "development flag will allow to run the operator outside a kubernetes cluster")
	flag.BoolVar(&c.debug, "debug", false, "enable debug mode")
	flag.StringVar(&c.listenAddr, "listen-address", ":9710", "Address to listen on for metrics.")
	flag.StringVar(&c.metricsPath, "metrics-path", "/metrics", "Path to serve the metrics.")

	// Parse flags
	flag.Parse()
}

func (c *cmdFlags) toRedisOperatorConfig() redisfailover.Config {
	return redisfailover.Config{
		Labels:        map[string]string{},
		ListenAddress: c.listenAddr,
		MetricsPath:   c.metricsPath,
	}
}

// Main is the  main runner.
type Main struct {
	flags     *cmdFlags
	k8sConfig rest.Config
	logger    log.Logger
	stopC     chan struct{}
}

// New returns a Main object.
func New(logger log.Logger) Main {
	// Init flags.
	flgs := &cmdFlags{}
	flgs.init()

	return Main{
		logger: logger,
		flags:  flgs,
	}
}

// Run execs the program.
func (m *Main) Run() error {
	// Create signal channels.
	m.stopC = make(chan struct{})
	errC := make(chan error)

	// Set correct logging.
	if m.flags.debug {
		m.logger.Set("debug")
		m.logger.Debugf("debug mode activated")
	}

	// Create the metrics client.
	metricsServer := metrics.NewPrometheusMetrics(m.flags.metricsPath, http.DefaultServeMux)

	// Serve metrics.
	go func() {
		log.Infof("Listening on %s for metrics exposure", m.flags.listenAddr)
		http.ListenAndServe(m.flags.listenAddr, nil)
	}()

	// Kubernetes clients.
	stdclient, customclient, aeClientset, err := m.createKubernetesClients()
	if err != nil {
		return err
	}

	// Create kubernetes service.
	k8sservice := k8s.New(stdclient, customclient, aeClientset, m.logger)

	// Create the redis clients
	redisClient := redis.New()

	// Create operator and run.
	redisfailoverOperator := redisfailover.New(m.flags.toRedisOperatorConfig(), k8sservice, redisClient, metricsServer, m.logger)
	go func() {
		errC <- redisfailoverOperator.Run(m.stopC)
	}()

	// Await signals.
	sigC := m.createSignalCapturer()
	var finalErr error
	select {
	case <-sigC:
		m.logger.Infof("Signal captured, exiting...")
	case err := <-errC:
		m.logger.Errorf("Error received: %s, exiting...", err)
		finalErr = err
	}

	m.stop(m.stopC)
	return finalErr
}

// loadKubernetesConfig loads kubernetes configuration based on flags.
func (m *Main) loadKubernetesConfig() (*rest.Config, error) {
	var cfg *rest.Config
	// If devel mode then use configuration flag path.
	if m.flags.development {
		config, err := clientcmd.BuildConfigFromFlags("", m.flags.kubeConfig)
		if err != nil {
			return nil, fmt.Errorf("could not load configuration: %s", err)
		}
		cfg = config
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("error loading kubernetes configuration inside cluster, check app is running outside kubernetes cluster or run in development mode: %s", err)
		}
		cfg = config
	}

	return cfg, nil
}

func (m *Main) createKubernetesClients() (kubernetes.Interface, redisfailoverclientset.Interface, apiextensionsclientset.Interface, error) {
	config, err := m.loadKubernetesConfig()
	if err != nil {
		return nil, nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}
	customClientset, err := redisfailoverclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	aeClientset, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	return clientset, customClientset, aeClientset, nil
}

func (m *Main) createSignalCapturer() <-chan os.Signal {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	return sigC
}

func (m *Main) stop(stopC chan struct{}) {
	m.logger.Infof("Stopping everything, waiting %s...", gracePeriod)

	// stop everything and let them time to stop
	close(stopC)
	time.Sleep(gracePeriod)
}

// Run app.
func main() {
	logger := log.Base()
	m := New(logger)

	if err := m.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error executing: %s", err)
		os.Exit(1)
	}
	os.Exit(0)
}
