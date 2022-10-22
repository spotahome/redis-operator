package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	redisfailoverclientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
)

// RedisFailover the RF service that knows how to interact with k8s to get them
type RedisFailover interface {
	// ListRedisFailovers lists the redisfailovers on a cluster.
	ListRedisFailovers(ctx context.Context, namespace string, opts metav1.ListOptions) (*redisfailoverv1.RedisFailoverList, error)
	// WatchRedisFailovers watches the redisfailovers on a cluster.
	WatchRedisFailovers(ctx context.Context, namespace string, opts metav1.ListOptions) (watch.Interface, error)
}

// RedisFailoverService is the RedisFailover service implementation using API calls to kubernetes.
type RedisFailoverService struct {
	k8sCli          redisfailoverclientset.Interface
	logger          log.Logger
	metricsRecorder metrics.Recorder
}

// NewRedisFailoverService returns a new Workspace KubeService.
func NewRedisFailoverService(k8scli redisfailoverclientset.Interface, logger log.Logger, metricsRecorder metrics.Recorder) *RedisFailoverService {
	logger = logger.With("service", "k8s.redisfailover")
	return &RedisFailoverService{
		k8sCli:          k8scli,
		logger:          logger,
		metricsRecorder: metricsRecorder,
	}
}

// ListRedisFailovers satisfies redisfailover.Service interface.
func (r *RedisFailoverService) ListRedisFailovers(ctx context.Context, namespace string, opts metav1.ListOptions) (*redisfailoverv1.RedisFailoverList, error) {
	redisFailoverList, err := r.k8sCli.DatabasesV1().RedisFailovers(namespace).List(ctx, opts)
	recordMetrics(namespace, "RedisFailover", metrics.NOT_APPLICABLE, "LIST", err, r.metricsRecorder)
	return redisFailoverList, err
}

// WatchRedisFailovers satisfies redisfailover.Service interface.
func (r *RedisFailoverService) WatchRedisFailovers(ctx context.Context, namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	watcher, err := r.k8sCli.DatabasesV1().RedisFailovers(namespace).Watch(ctx, opts)
	recordMetrics(namespace, "RedisFailover", metrics.NOT_APPLICABLE, "WATCH", err, r.metricsRecorder)
	return watcher, err
}
