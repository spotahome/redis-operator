package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/spotahome/redis-operator/cmd/utils"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	"github.com/spotahome/redis-operator/operator/redisfailover"
	"github.com/spotahome/redis-operator/service/k8s"
	"github.com/spotahome/redis-operator/service/redis"
)

const (
	gracePeriod      = 5 * time.Second
	metricsNamespace = "redis_operator"
)

// Main is the  main runner.
type Main struct {
	flags  *utils.CMDFlags
	logger log.Logger
	stopC  chan struct{}
}

// New returns a Main object.
func New(logger log.Logger) Main {
	// Init flags.
	flgs := &utils.CMDFlags{}
	flgs.Init()

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
	if m.flags.Debug {
		err := m.logger.Set("debug")
		if err != nil {
			return err
		}
		m.logger.Debugf("debug mode activated")
	}

	// Create the metrics client.
	metricsRecorder := metrics.NewRecorder(metricsNamespace, prometheus.DefaultRegisterer)

	// Serve metrics.
	go func() {
		log.Infof("Listening on %s for metrics exposure on URL %s", m.flags.ListenAddr, m.flags.MetricsPath)
		http.Handle(m.flags.MetricsPath, promhttp.Handler())
		err := http.ListenAndServe(m.flags.ListenAddr, nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Kubernetes clients.
	stdclient, customclient, aeClientset, err := utils.CreateKubernetesClients(m.flags)
	if err != nil {
		return err
	}

	// Create kubernetes service.
	k8sservice := k8s.New(stdclient, customclient, aeClientset, m.logger)

	// Create the redis clients
	redisClient := redis.New()

	// Create operator and run.
	redisfailoverOperator, err := redisfailover.New(m.flags.ToRedisOperatorConfig(), k8sservice, redisClient, metricsRecorder, m.logger)
	if err != nil {
		return err
	}

	go func() {
		errC <- redisfailoverOperator.Run(context.Background())
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
