package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/monitoring/metrics"
	"github.com/spotahome/kooper/operator/controller"
	"github.com/spotahome/kooper/operator/handler"
	"github.com/spotahome/kooper/operator/retrieve"
)

const (
	metricsPrefix     = "metricsexample"
	metricsAddr       = ":7777"
	prometheusBackend = "prometheus"
)

var (
	metricsBackend string
)

func initFlags() error {
	fg := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fg.StringVar(&metricsBackend, "metrics-backend", "prometheus", "the metrics backend to use")
	return fg.Parse(os.Args[1:])
}

// sleep will sleep randomly from 0 to 1000 milliseconds.
func sleepRandomly() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sleepMS := r.Intn(10) * 100
	time.Sleep(time.Duration(sleepMS) * time.Millisecond)
}

// errRandomly will will err randomly.
func errRandomly() error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Intn(10)%3 == 0 {
		return fmt.Errorf("random error :)")
	}
	return nil
}

// creates prometheus recorder and starts serving metrics in background.
func createPrometheusRecorder(logger log.Logger) metrics.Recorder {
	// We could use also prometheus global registry (the default one)
	// prometheus.DefaultRegisterer instead of creating a new one
	reg := prometheus.NewRegistry()
	m := metrics.NewPrometheus(reg)

	// Start serving metrics in background.
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	go func() {
		logger.Infof("serving metrics at %s", metricsAddr)
		http.ListenAndServe(metricsAddr, h)
	}()

	return m
}

func getMetricRecorder(backend string, logger log.Logger) (metrics.Recorder, error) {
	switch backend {
	case prometheusBackend:
		logger.Infof("using Prometheus metrics recorder")
		return createPrometheusRecorder(logger), nil
	}

	return nil, fmt.Errorf("wrong metrics backend")
}

func main() {
	// Initialize logger.
	log := &log.Std{}

	// Init flags.
	if err := initFlags(); err != nil {
		log.Errorf("error parsing arguments: %s", err)
		os.Exit(1)
	}

	// Get k8s client.
	k8scfg, err := rest.InClusterConfig()
	if err != nil {
		// No in cluster? letr's try locally
		kubehome := filepath.Join(homedir.HomeDir(), ".kube", "config")
		k8scfg, err = clientcmd.BuildConfigFromFlags("", kubehome)
		if err != nil {
			log.Errorf("error loading kubernetes configuration: %s", err)
			os.Exit(1)
		}
	}
	k8scli, err := kubernetes.NewForConfig(k8scfg)
	if err != nil {
		log.Errorf("error creating kubernetes client: %s", err)
		os.Exit(1)
	}

	// Create our retriever so the controller knows how to get/listen for pod events.
	retr := &retrieve.Resource{
		Object: &corev1.Pod{},
		ListerWatcher: &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return k8scli.CoreV1().Pods("").List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return k8scli.CoreV1().Pods("").Watch(options)
			},
		},
	}

	// Our domain logic that will print every add/sync/update and delete event we .
	hand := &handler.HandlerFunc{
		AddFunc: func(_ context.Context, obj runtime.Object) error {
			sleepRandomly()
			return errRandomly()
		},
		DeleteFunc: func(_ context.Context, s string) error {
			sleepRandomly()
			return errRandomly()
		},
	}

	// Create the controller that will refresh every 30 seconds.
	m, err := getMetricRecorder(metricsBackend, log)
	if err != nil {
		log.Errorf("errors getting metrics backend: %s", err)
		os.Exit(1)
	}
	cfg := &controller.Config{
		Name: "metricsControllerTest",
	}
	ctrl := controller.New(cfg, hand, retr, nil, nil, m, log)

	// Start our controller.
	stopC := make(chan struct{})
	if err := ctrl.Run(stopC); err != nil {
		log.Errorf("error running controller: %s", err)
		os.Exit(1)
	}
	os.Exit(0)
}
