package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	koopercontroller "github.com/spotahome/kooper/v2/controller"
	kooperprometheus "github.com/spotahome/kooper/v2/metrics/prometheus"
)

const (
	promControllerSubsystem = "controller"
)

const ()

// Instrumenter is the interface that will collect the metrics and has ability to send/expose those metrics.
type Recorder interface {
	koopercontroller.MetricsRecorder

	// ClusterOK metrics
	SetClusterOK(namespace string, name string)
	SetClusterError(namespace string, name string)
	DeleteCluster(namespace string, name string)

	// Indicate if `ensure` operation succeeded
	IncrEnsureResourceSuccessCount(objectNamespace string, objectName string, objectKind string, resourceName string)
	// Indicate if `ensure` operation failed
	IncrEnsureResourceFailureCount(objectNamespace string, objectName string, objectKind string, resourceName string)
	// Indicate redis instances being monitored
	SetRedisInstance(IP string, masterIP string, role string)
}

// PromMetrics implements the instrumenter so the metrics can be managed by Prometheus.
type recorder struct {
	// Metrics fields.
	clusterOK             *prometheus.GaugeVec   // clusterOk is the status of a cluster
	ensureResourceSuccess *prometheus.CounterVec // number of successful "ensure" operators performed by the controller.
	ensureResourceFailure *prometheus.CounterVec // number of failed "ensure" operators performed by the controller.
	redisInstance         *prometheus.GaugeVec   // Indicates known redis instances, with IPs and master/slave status
	koopercontroller.MetricsRecorder
}

type ensureResourceSuccessRecorder struct {
	ensureResourceSuccess *prometheus.CounterVec
	koopercontroller.MetricsRecorder
}

type ensureResourceFailureRecorder struct {
	ensureResourceSuccess *prometheus.CounterVec
	koopercontroller.MetricsRecorder
}

type redisInstanceRecorder struct {
	ensureResourceSuccess *prometheus.GaugeVec
	koopercontroller.MetricsRecorder
}

// NewPrometheusMetrics returns a new PromMetrics object.
func NewRecorder(namespace string, reg prometheus.Registerer) Recorder {
	// Create metrics.
	clusterOK := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: promControllerSubsystem,
		Name:      "cluster_ok",
		Help:      "Number of failover clusters managed by the operator.",
	}, []string{"namespace", "name"})

	ensureResourceSuccess := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: promControllerSubsystem,
		Name:      "ensure_resource_success",
		Help:      "number of successful 'ensure' operations on a resource performed by the controller.",
	}, []string{"object_namespace", "object_name", "object_kind", "resource_name"})

	ensureResourceFailure := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: promControllerSubsystem,
		Name:      "ensure_resource_failure",
		Help:      "number of failed 'ensure' operations on a resource performed by the controller.",
	}, []string{"object_namespace", "object_name", "object_kind", "resource_name"})

	redisInstance := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: promControllerSubsystem,
		Name:      "redis_instance",
		Help:      "redis instances discovered guage. IPs of redis instances, and Master/Slave role as indicators in the labels.",
	}, []string{"IP", "MasterIP", "role"})

	// Create the instance.
	r := recorder{
		clusterOK:             clusterOK,
		ensureResourceSuccess: ensureResourceSuccess,
		ensureResourceFailure: ensureResourceFailure,
		redisInstance:         redisInstance,
		MetricsRecorder: kooperprometheus.New(kooperprometheus.Config{
			Registerer: reg,
		}),
	}

	// Register metrics.
	reg.MustRegister(
		r.clusterOK,
		r.ensureResourceSuccess,
		r.ensureResourceFailure,
		r.redisInstance,
	)

	return r
}

// SetClusterOK set the cluster status to OK
func (r recorder) SetClusterOK(namespace string, name string) {
	r.clusterOK.WithLabelValues(namespace, name).Set(1)
}

// SetClusterError set the cluster status to Error
func (r recorder) SetClusterError(namespace string, name string) {
	r.clusterOK.WithLabelValues(namespace, name).Set(0)
}

// DeleteCluster set the cluster status to Error
func (r recorder) DeleteCluster(namespace string, name string) {
	r.clusterOK.DeleteLabelValues(namespace, name)
}

func (r recorder) IncrEnsureResourceSuccessCount(objectNamespace string, objectName string, objectKind string, resourceName string) {
	r.ensureResourceSuccess.WithLabelValues(objectNamespace, objectName, objectKind, resourceName).Add(1)
}

func (r recorder) IncrEnsureResourceFailureCount(objectNamespace string, objectName string, objectKind string, resourceName string) {
	r.ensureResourceSuccess.WithLabelValues(objectNamespace, objectName, objectKind, resourceName).Add(1)
}

func (r recorder) SetRedisInstance(IP string, masterIP string, role string) {
	r.redisInstance.WithLabelValues(IP, masterIP, role).Set(1)
}
