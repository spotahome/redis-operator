package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	promNamespace           = "redis_operator"
	promControllerSubsystem = "controller"
	promClusterSubsystem    = "cluster"

	eventAdd    = "add"
	eventUpdate = "update"
	eventDelete = "delete"

	stateRunning  = "running"
	stateCreating = "creating"
	stateFailed   = "failed"
)

// Instrumenter is the interface that will collect the metrics and has ability to send/expose those metrics.
type Instrumenter interface {
	// SetClustersRunning sets the number of running clusters managed by the controller in the current state.
	SetClustersRunning(n float64)
	// SetClustersRunning sets the number of creating clusters managed by the controller in the current state.
	SetClustersCreating(n float64)
	// SetClustersRunning sets the number of failed clusters managed by the controller in the current state.
	SetClustersFailed(n float64)
	// IncIncAddEventHandled increments the number of add events handled by the controller.
	IncAddEventHandled(cluster string)
	// IncIncUpdateEventHandled increments the number of update events handled by the controller.
	IncUpdateEventHandled(cluster string)
	// IncIncDeleteEventHandled increments the number of delete events handled by the controller.
	IncDeleteEventHandled(cluster string)
	// SetClusterMasters sets the number of masters a cluster has.
	SetClusterMasters(n float64, cluster string)
}

// PromMetrics implements the instrumenter so the metrics can be managed by Prometheus.
type PromMetrics struct {
	// Metrics fields.
	clusterTotal        *prometheus.GaugeVec   // clusterTotal is the total number of clusters managed by the operator.
	eventTotal          *prometheus.CounterVec // eventTotal is the total number of k8s events received & handled by the operator.
	clusterMastersTotal *prometheus.GaugeVec   // clusterMasterTotal is the total number of masters a cluster has.

	// Instrumentation fields.
	registry prometheus.Registerer
	path     string
	mux      *http.ServeMux
}

// NewPrometheusMetrics returns a new PromMetrics object.
func NewPrometheusMetrics(path string, mux *http.ServeMux) *PromMetrics {
	// Create metrics.
	clusterTotal := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promControllerSubsystem,
		Name:      "clusters",
		Help:      "Number of failover clusters managed by the operator.",
	}, []string{"state"})

	eventTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promControllerSubsystem,
		Name:      "event_handled_total",
		Help:      "Cumulative number of K8s events handled by the operator.",
	}, []string{"kind", "cluster"})

	clusterMastersTotal := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promClusterSubsystem,
		Name:      "masters",
		Help:      "Number of masters a failover clusters has.",
	}, []string{"cluster"})

	// Create Prometheus registry, use this instead of the prometheus default registry.
	// TODO: Do we need go default instrumentation bout the go process metrics?
	promReg := prometheus.NewRegistry()

	// Create the instance.
	p := &PromMetrics{
		clusterTotal:        clusterTotal,
		eventTotal:          eventTotal,
		clusterMastersTotal: clusterMastersTotal,

		registry: promReg,
		path:     path,
		mux:      mux,
	}

	// Register metrics on prometheus.
	p.register()

	// Register prometheus handler so we can serve the metrics.
	handler := promhttp.HandlerFor(promReg, promhttp.HandlerOpts{})
	mux.Handle(path, handler)

	return p
}

// register will register all the required prometheus metrics on the Prometheus collector.
func (p *PromMetrics) register() {
	p.registry.MustRegister(p.clusterTotal)
	p.registry.MustRegister(p.eventTotal)
	p.registry.MustRegister(p.clusterMastersTotal)
}

// SetClustersRunning satisfies Instrumenter interface.
func (p *PromMetrics) SetClustersRunning(n float64) {
	p.clusterTotal.WithLabelValues(stateRunning).Set(n)
}

// SetClustersCreating satisfies Instrumenter interface.
func (p *PromMetrics) SetClustersCreating(n float64) {
	p.clusterTotal.WithLabelValues(stateCreating).Set(n)
}

// SetClustersFailed satisfies Instrumenter interface.
func (p *PromMetrics) SetClustersFailed(n float64) {
	p.clusterTotal.WithLabelValues(stateFailed).Set(n)
}

// IncAddEventHandled satisfies Instrumenter interface.
func (p *PromMetrics) IncAddEventHandled(cluster string) {
	p.eventTotal.WithLabelValues(eventAdd, cluster).Inc()
}

// IncUpdateEventHandled satisfies Instrumenter interface.
func (p *PromMetrics) IncUpdateEventHandled(cluster string) {
	p.eventTotal.WithLabelValues(eventUpdate, cluster).Inc()
}

// IncDeleteEventHandled satisfies Instrumenter interface.
func (p *PromMetrics) IncDeleteEventHandled(cluster string) {
	p.eventTotal.WithLabelValues(eventDelete, cluster).Inc()
}

// SetClusterMasters satisfies Instrumenter interface.
func (p *PromMetrics) SetClusterMasters(n float64, cluster string) {
	p.clusterMastersTotal.WithLabelValues(cluster).Set(n)
}
