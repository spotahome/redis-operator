package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	promControllerSubsystem = "controller"
)

// Instrumenter is the interface that will collect the metrics and has ability to send/expose those metrics.
type Instrumenter interface {
	SetClusterOK(namespace string, name string)
	SetClusterError(namespace string, name string)
	DeleteCluster(namespace string, name string)
}

// PromMetrics implements the instrumenter so the metrics can be managed by Prometheus.
type PromMetrics struct {
	// Metrics fields.
	clusterOK *prometheus.GaugeVec // clusterOk is the status of a cluster

	// Instrumentation fields.
	registry prometheus.Registerer
	path     string
	mux      *http.ServeMux
}

// NewPrometheusMetrics returns a new PromMetrics object.
func NewPrometheusMetrics(path string, namespace string, mux *http.ServeMux, registry *prometheus.Registry) *PromMetrics {
	// Create metrics.
	clusterOK := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: promControllerSubsystem,
		Name:      "cluster_ok",
		Help:      "Number of failover clusters managed by the operator.",
	}, []string{"namespace", "name"})

	// Create the instance.
	p := &PromMetrics{
		clusterOK: clusterOK,

		registry: registry,
		path:     path,
		mux:      mux,
	}

	// Register metrics on prometheus.
	p.register()

	// Register prometheus handler so we can serve the metrics.
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	mux.Handle(path, handler)

	return p
}

// register will register all the required prometheus metrics on the Prometheus collector.
func (p *PromMetrics) register() {
	p.registry.MustRegister(p.clusterOK)
}

// SetClusterOK set the cluster status to OK
func (p *PromMetrics) SetClusterOK(namespace string, name string) {
	p.clusterOK.WithLabelValues(namespace, name).Set(1)
}

// SetClusterError set the cluster status to Error
func (p *PromMetrics) SetClusterError(namespace string, name string) {
	p.clusterOK.WithLabelValues(namespace, name).Set(0)
}

// DeleteCluster set the cluster status to Error
func (p *PromMetrics) DeleteCluster(namespace string, name string) {
	p.clusterOK.DeleteLabelValues(namespace, name)
}
