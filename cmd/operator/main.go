package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spotahome/redis-operator/pkg/failover"
	"github.com/spotahome/redis-operator/pkg/log"
	"github.com/spotahome/redis-operator/pkg/metrics"
)

type flags struct {
	kubeConfig  string
	listenAddr  string
	metricsPath string
	maxThreads  int
}

func initFlags() flags {
	f := flags{}

	// Register flags.
	flag.StringVar(&f.kubeConfig, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	flag.StringVar(&f.listenAddr, "listen-address", ":9710", "Address to listen on for metrics.")
	flag.StringVar(&f.metricsPath, "metrics-path", "/metrics", "Path to serve the metrics.")
	flag.IntVar(&f.maxThreads, "max-threads", 0, "Maximum number of threads.")

	// Parse flags & return.
	flag.Parse()
	return f
}

func main() {
	log.Set(log.Level("debug"))

	log.Info("Redis-Operator Starting...")
	flags := initFlags()

	// Create the metrics client.
	m := metrics.NewPrometheusMetrics(flags.metricsPath, http.DefaultServeMux)

	// Serve metrics.
	go func() {
		log.Infof("Listening on %s", flags.listenAddr)
		http.ListenAndServe(flags.listenAddr, nil)
	}()

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(flags.kubeConfig)
	if err != nil {
		panic(err)
	}

	// Create the clientset for accessing Kubernetes API
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	if flags.maxThreads == 0 {
		flags.maxThreads = runtime.GOMAXPROCS(0)
	}

	stop := make(chan int, 1)
	crd, err := failover.NewRedisFailoverCRD(m, clientset, *config, stop, flags.maxThreads)
	if err != nil {
		log.Panic(err)
	}

	// Initialize third party resource if it does not exist
	err = crd.Create()
	if err != nil && !apierrors.IsAlreadyExists(err) {
		panic(err)

	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	crd.Watch()
	<-c
	stop <- 1
	log.Info("Exiting...")
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
