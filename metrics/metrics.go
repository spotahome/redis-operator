package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	koopercontroller "github.com/spotahome/kooper/v2/controller"
	kooperprometheus "github.com/spotahome/kooper/v2/metrics/prometheus"
)

const (
	promControllerSubsystem = "controller"
)

// variables for setting various indicator labels
const (
	NOT_APPLICABLE            = "NA"
	UNHEALTHY                 = 1.0
	HEALTHY                   = 0.0
	REDIS_REPLICA_MISMATCH    = "REDIS_STATEFULSET_REPLICAS_MISMATCH"
	SENTINEL_REPLICA_MISMATCH = "SENTINEL_DEPLOYMENT_REPLICAS_MISMATCH"
	NUMBER_OF_MASTERS         = "MASTER_COUNT_IS_NOT_ONE"
	SENTINEL_WRONG_MASTER     = "SENTINEL_IS_CONFIGURED_WITH_WRONG_MASTER_IP"
	SLAVE_WRONG_MASTER        = "SLAVE_IS_CONFIGURED_WITH_WRONG_MASTER_IP"

	// redis connection related errors
	WRONG_PASSWORD_USED = "WRONG_PASSWORD_USED"
	NOAUTH              = "AUTH_CREDENTIALS_NOT_PROVIDED"
	NOPERM              = "REDIS_USER_DOES_NOT_HAVE_PERMISSIONS"
	IO_TIMEOUT          = "CONNECTION_TIMEDOUT"
	CONNECTION_REFUSED  = "CONNECTION_REFUSED"
)

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
	ResetRedisInstance()

	IncrRedisUnhealthyCount(namespace string, resource string, indicator /* aspect of redis that is unhealthy */ string, instance string)
	IncrSentinelUnhealthyCount(namespace string, resource string, indicator /* aspect of redis that is unhealthy */ string, instance string)
}

// PromMetrics implements the instrumenter so the metrics can be managed by Prometheus.
type recorder struct {
	// Metrics fields.
	clusterOK             *prometheus.GaugeVec   // clusterOk is the status of a cluster
	ensureResourceSuccess *prometheus.CounterVec // number of successful "ensure" operators performed by the controller.
	ensureResourceFailure *prometheus.CounterVec // number of failed "ensure" operators performed by the controller.
	redisInstance         *prometheus.GaugeVec   // Indicates known redis instances, with IPs and master/slave status
	redisUnhealthy        *prometheus.CounterVec // indicates any error encountered in managed redis instance(s)
	sentinelUnhealthy     *prometheus.CounterVec // indicates any error encountered in managed sentinel instance(s)
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

type redisHealthRecorder struct {
	redisUnhealthy *prometheus.CounterVec
	koopercontroller.MetricsRecorder
}

type sentinelHealthRecorder struct {
	sentinelUnhealthy *prometheus.CounterVec
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
		Name:      "redis_instance_info",
		Help:      "redis instances discovered. IPs of redis instances, and Master/Slave role as indicators in the labels.",
	}, []string{"IP", "MasterIP", "role"})

	redisUnhealthy := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: promControllerSubsystem,
		Name:      "redis_unhealthy",
		Help:      "indicates any error encountered in managed redis instance(s)",
	}, []string{"namespace", "resource", "indicator", "instance"})

	sentinelUnhealthy := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: promControllerSubsystem,
		Name:      "sentinel_unhealthy",
		Help:      "indicates any error encountered in managed sentinel instance(s)",
	}, []string{"namespace", "resource", "indicator", "instance"})

	// Create the instance.
	r := recorder{
		clusterOK:             clusterOK,
		ensureResourceSuccess: ensureResourceSuccess,
		ensureResourceFailure: ensureResourceFailure,
		redisInstance:         redisInstance,
		redisUnhealthy:        redisUnhealthy,
		sentinelUnhealthy:     sentinelUnhealthy,
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
		r.redisUnhealthy,
		r.sentinelUnhealthy,
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

func (r recorder) ResetRedisInstance() {
	r.redisInstance.Reset()
}

func (r recorder) IncrRedisUnhealthyCount(namespace string, resource string, indicator /* aspect of redis that is unhealthy */ string, instance string) {
	r.redisUnhealthy.WithLabelValues(namespace, resource, indicator, instance).Add(1)
}

func (r recorder) IncrSentinelUnhealthyCount(namespace string, resource string, indicator /* aspect of sentinel that is unhealthy */ string, instance string) {
	r.sentinelUnhealthy.WithLabelValues(namespace, resource, indicator, instance).Add(1)
}
