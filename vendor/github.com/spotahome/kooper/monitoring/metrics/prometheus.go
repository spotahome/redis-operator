package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	promNamespace           = "kooper"
	promControllerSubsystem = "controller"
)

// Prometheus implements the metrics recording in a prometheus registry.
type Prometheus struct {
	// Metrics
	queuedEvents           *prometheus.CounterVec
	processedEvents        *prometheus.CounterVec
	processedEventErrors   *prometheus.CounterVec
	processedEventDuration *prometheus.HistogramVec

	reg prometheus.Registerer
}

// NewPrometheus returns a new Prometheus metrics backend with metrics prefixed by the namespace.
func NewPrometheus(registry prometheus.Registerer) *Prometheus {
	return NewPrometheusWithBuckets(prometheus.DefBuckets, registry)
}

// NewPrometheusWithBuckets returns a new Prometheus metrics backend with metrics prefixed by the
// namespace and with custom buckets for the duration/latency metrics. This kind should be used when
// the default buckets don't work. This could happen when the time to process an event is not on the
// range of 5ms-10s duration.
// Check https://godoc.org/github.com/prometheus/client_golang/prometheus#pkg-variables
func NewPrometheusWithBuckets(buckets []float64, registry prometheus.Registerer) *Prometheus {
	p := &Prometheus{
		queuedEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: promNamespace,
			Subsystem: promControllerSubsystem,
			Name:      "queued_events_total",
			Help:      "Total number of events queued.",
		}, []string{"controller", "type"}),

		processedEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: promNamespace,
			Subsystem: promControllerSubsystem,
			Name:      "processed_events_total",
			Help:      "Total number of successfuly processed events.",
		}, []string{"controller", "type"}),

		processedEventErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: promNamespace,
			Subsystem: promControllerSubsystem,
			Name:      "processed_event_errors_total",
			Help:      "Total number of errors processing events.",
		}, []string{"controller", "type"}),

		processedEventDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: promNamespace,
			Subsystem: promControllerSubsystem,
			Name:      "processed_event_duration_seconds",
			Help:      "The duration for a successful event to be processed.",
			Buckets:   buckets,
		}, []string{"controller", "type"}),
		reg: registry,
	}

	p.registerMetrics()
	return p
}

func (p *Prometheus) registerMetrics() {
	p.reg.MustRegister(
		p.queuedEvents,
		p.processedEvents,
		p.processedEventErrors,
		p.processedEventDuration)

}

// IncResourceEventQueued satisfies metrics.Recorder interface.
func (p *Prometheus) IncResourceEventQueued(controller string, eventType EventType) {
	p.queuedEvents.WithLabelValues(controller, string(eventType)).Inc()
}

// IncResourceEventProcessed satisfies metrics.Recorder interface.
func (p *Prometheus) IncResourceEventProcessed(controller string, eventType EventType) {
	p.processedEvents.WithLabelValues(controller, string(eventType)).Inc()
}

// IncResourceEventProcessedError satisfies metrics.Recorder interface.
func (p *Prometheus) IncResourceEventProcessedError(controller string, eventType EventType) {
	p.processedEventErrors.WithLabelValues(controller, string(eventType)).Inc()
}

// ObserveDurationResourceEventProcessed satisfies metrics.Recorder interface.
func (p *Prometheus) ObserveDurationResourceEventProcessed(controller string, eventType EventType, start time.Time) {
	secs := time.Now().Sub(start).Seconds()
	p.processedEventDuration.WithLabelValues(controller, string(eventType)).Observe(secs)
}
