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
	SUCCESS                                = "SUCCESS"
	FAIL                                   = "FAIL"
	NOT_APPLICABLE                         = "NA"
	UNHEALTHY                              = 1.0
	HEALTHY                                = 0.0
	REDIS_REPLICA_MISMATCH                 = "REDIS_STATEFULSET_REPLICAS_MISMATCH"
	SENTINEL_REPLICA_MISMATCH              = "SENTINEL_DEPLOYMENT_REPLICAS_MISMATCH"
	NUMBER_OF_MASTERS                      = "MASTER_COUNT_IS_NOT_ONE"
	SENTINEL_WRONG_MASTER                  = "SENTINEL_IS_CONFIGURED_WITH_WRONG_MASTER_IP"
	SLAVE_WRONG_MASTER                     = "SLAVE_IS_CONFIGURED_WITH_WRONG_MASTER_IP"
	SENTINEL_NOT_READY                     = "SENTINEL_NOT_READY"
	REGEX_NOT_FOUND                        = "SENTINEL_REGEX_NOT_FOUND"
	MISC                                   = "MISC_ERROR"
	SENTINEL_NUMBER_IN_MEMORY_MISMATCH     = "SENTINEL_NUMBER_IN_MEMORY_MISMATCH"
	REDIS_SLAVES_NUMBER_IN_MEMORY_MISMATCH = "REDIS_SLAVES_NUMBER_IN_MEMORY_MISMATCH"
	// redis connection related errors
	WRONG_PASSWORD_USED = "WRONG_PASSWORD_USED"
	NOAUTH              = "AUTH_CREDENTIALS_NOT_PROVIDED"
	NOPERM              = "REDIS_USER_DOES_NOT_HAVE_PERMISSIONS"
	IO_TIMEOUT          = "CONNECTION_TIMEDOUT"
	CONNECTION_REFUSED  = "CONNECTION_REFUSED"

	K8S_FORBIDDEN_ERR = "USER_FORBIDDEN_TO_PERFORM_ACTION"
	K8S_UNAUTH        = "CLIENT_NOT_AUTHORISED"
	K8S_MISC          = "MISC_ERROR_CHECK_LOGS"
	K8S_NOT_FOUND     = "RESOURCE_NOT_FOUND"

	KIND_REDIS                  = "REDIS"
	KIND_SENTINEL               = "SENTINEL"
	APPLY_REDIS_CONFIG          = "APPLY_REDIS_CONFIG"
	APPLY_SENTINEL_CONFIG       = "APPLY_SENTINEL_CONFIG"
	MONITOR_REDIS_WITH_PORT     = "SET_SENTINEL_TO_MONITOR_REDIS_WITH_GIVEN_PORT"
	RESET_SENTINEL              = "RESET_ALL_SENTINEL_CONFIG"
	GET_NUM_SENTINELS_IN_MEM    = "GET_NUMBER_OF_SENTINELS_IN_MEMORY"    // `info sentinel` command on a sentinel machine > grep sentinel
	GET_NUM_REDIS_SLAVES_IN_MEM = "GET_NUMBER_OF_REDIS_SLAVES_IN_MEMORY" // `info sentinel` command on a sentinel machine > grep slaves
	GET_SLAVE_OF                = "GET_MASTER_OF_GIVEN_SLAVE_INSTANCE"
	IS_MASTER                   = "CHECK_IF_INSTANCE_IS_MASTER"
	MAKE_MASTER                 = "MAKE_INSTANCE_AS_MASTER"
	MAKE_SLAVE_OF               = "MAKE_SLAVE_OF_GIVEN_MASTER_INSTANCE"
	GET_SENTINEL_MONITOR        = "SENTINEL_GET_MASTER_INSTANCE"
	SLAVE_IS_READY              = "CHECK_IF_SLAVE_IS_READY"
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

	RecordK8sOperation(namespace string, kind string, object string, operation string, status string, err string)

	RecordRedisOperation(kind /*redis/sentinel? */ string, IP string, operation string, status string, err string)
}

// PromMetrics implements the instrumenter so the metrics can be managed by Prometheus.
type recorder struct {
	// Metrics fields.
	clusterOK             *prometheus.GaugeVec   // clusterOk is the status of a cluster
	ensureResourceSuccess *prometheus.CounterVec // number of successful "ensure" operators performed by the controller.
	ensureResourceFailure *prometheus.CounterVec // number of failed "ensure" operators performed by the controller.
	redisInstance         *prometheus.GaugeVec   // indicates known redis instances, with IPs and master/slave status
	redisUnhealthy        *prometheus.CounterVec // indicates any error encountered in managed redis instance(s)
	sentinelUnhealthy     *prometheus.CounterVec // indicates any error encountered in managed sentinel instance(s)
	k8sServiceOperations  *prometheus.CounterVec // number of operations performed on k8s
	redisOperations       *prometheus.CounterVec // number of operations performed on redis/sentinel instances
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

type k8sServiceOperationsRecorder struct {
	operations *prometheus.CounterVec
	koopercontroller.MetricsRecorder
}

type k8sServiceerrorRecorder struct {
	errors *prometheus.CounterVec
	koopercontroller.MetricsRecorder
}

type redisHealthRecorder struct {
	redisUnhealthy *prometheus.CounterVec
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

	redisOperations := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: promControllerSubsystem,
			Name:      "redis_operations",
			Help:      "number of operations performed on redis",
		}, []string{"kind" /* redis/sentinel? */, "IP", "operation", "status", "err"})

	k8sServiceOperations := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: promControllerSubsystem,
			Name:      "k8s_operations",
			Help:      "number of operations performed on k8s",
		}, []string{"namespace", "kind", "object", "operation", "status", "err"})
	// Create the instance.
	r := recorder{
		clusterOK:             clusterOK,
		ensureResourceSuccess: ensureResourceSuccess,
		ensureResourceFailure: ensureResourceFailure,
		redisInstance:         redisInstance,
		redisUnhealthy:        redisUnhealthy,
		sentinelUnhealthy:     sentinelUnhealthy,
		k8sServiceOperations:  k8sServiceOperations,
		redisOperations:       redisOperations,
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
		r.k8sServiceOperations,
		r.redisOperations,
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

func (r recorder) RecordK8sOperation(namespace string, kind string, object string, operation string, status string, err string) {
	r.k8sServiceOperations.WithLabelValues(namespace, kind, object, operation, status, err).Add(1)
}

func (r recorder) RecordRedisOperation(kind /*redis/sentinel? */ string, IP string, operation string, status string, err string) {
	r.redisOperations.WithLabelValues(kind, IP, operation, status, err).Add(1)
}
