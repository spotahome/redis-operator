package prometheus

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spotahome/kooper/v2/controller"
)

const (
	promNamespace           = "kooper"
	promControllerSubsystem = "controller"
)

// Config is the Recorder Config.
type Config struct {
	// Registerer is a prometheus registerer, e.g: prometheus.Registry.
	// By default will use Prometheus default registry.
	Registerer prometheus.Registerer
	// InQueueBuckets sets custom buckets for the duration/latency items in queue metrics.
	// Check https://godoc.org/github.com/prometheus/client_golang/prometheus#pkg-variables
	InQueueBuckets []float64
	// ProcessingBuckets sets custom buckets for the duration/latency processing metrics.
	// Check https://godoc.org/github.com/prometheus/client_golang/prometheus#pkg-variables
	ProcessingBuckets []float64
}

func (c *Config) defaults() {
	if c.Registerer == nil {
		c.Registerer = prometheus.DefaultRegisterer
	}

	if c.InQueueBuckets == nil || len(c.InQueueBuckets) == 0 {
		// Use bigger buckets thant he default ones because the times of waiting queues
		// usually are greater than the handling, and resync of events can be minutes.
		c.InQueueBuckets = []float64{.01, .05, .1, .25, .5, 1, 3, 10, 20, 60, 150, 300}
	}

	if c.ProcessingBuckets == nil || len(c.ProcessingBuckets) == 0 {
		c.ProcessingBuckets = prometheus.DefBuckets
	}
}

// Recorder implements the metrics recording in a prometheus registry.
type Recorder struct {
	reg prometheus.Registerer

	queuedEventsTotal      *prometheus.CounterVec
	inQueueEventDuration   *prometheus.HistogramVec
	processedEventDuration *prometheus.HistogramVec
}

// New returns a new Prometheus implementaiton for a metrics recorder.
func New(cfg Config) *Recorder {
	cfg.defaults()

	r := &Recorder{
		reg: cfg.Registerer,

		queuedEventsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: promNamespace,
			Subsystem: promControllerSubsystem,
			Name:      "queued_events_total",
			Help:      "Total number of events queued.",
		}, []string{"controller", "requeue"}),

		inQueueEventDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: promNamespace,
			Subsystem: promControllerSubsystem,
			Name:      "event_in_queue_duration_seconds",
			Help:      "The duration of an event in the queue.",
			Buckets:   cfg.InQueueBuckets,
		}, []string{"controller"}),

		processedEventDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: promNamespace,
			Subsystem: promControllerSubsystem,
			Name:      "processed_event_duration_seconds",
			Help:      "The duration for an event to be processed.",
			Buckets:   cfg.ProcessingBuckets,
		}, []string{"controller", "success"}),
	}

	// Register metrics.
	r.reg.MustRegister(
		r.queuedEventsTotal,
		r.inQueueEventDuration,
		r.processedEventDuration)

	return r
}

// IncResourceEventQueued satisfies controller.MetricsRecorder interface.
func (r Recorder) IncResourceEventQueued(ctx context.Context, controller string, isRequeue bool) {
	r.queuedEventsTotal.WithLabelValues(controller, strconv.FormatBool(isRequeue)).Inc()
}

// ObserveResourceInQueueDuration satisfies controller.MetricsRecorder interface.
func (r Recorder) ObserveResourceInQueueDuration(ctx context.Context, controller string, queuedAt time.Time) {
	r.inQueueEventDuration.WithLabelValues(controller).
		Observe(time.Since(queuedAt).Seconds())
}

// ObserveResourceProcessingDuration satisfies controller.MetricsRecorder interface.
func (r Recorder) ObserveResourceProcessingDuration(ctx context.Context, controller string, success bool, startProcessingAt time.Time) {
	r.processedEventDuration.WithLabelValues(controller, strconv.FormatBool(success)).
		Observe(time.Since(startProcessingAt).Seconds())
}

// RegisterResourceQueueLengthFunc satisfies controller.MetricsRecorder interface.
func (r Recorder) RegisterResourceQueueLengthFunc(controller string, f func(context.Context) int) error {
	err := r.reg.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   promNamespace,
			Subsystem:   promControllerSubsystem,
			Name:        "event_queue_length",
			Help:        "Length of the controller resource queue.",
			ConstLabels: prometheus.Labels{"controller": controller},
		},
		func() float64 { return float64(f(context.Background())) },
	))
	if err != nil {
		return fmt.Errorf("could not register ResourceQueueLengthFunc metrics: %w", err)
	}

	return nil
}

// Check interfaces implementation.
var _ controller.MetricsRecorder = &Recorder{}
