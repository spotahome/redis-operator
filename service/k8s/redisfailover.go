package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	redisfailoverclientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned"
	"github.com/spotahome/redis-operator/log"
)

// RedisFailover the RF service that knows how to interact with k8s to get them
type RedisFailover interface {
	// ListRedisFailovers lists the redisfailovers on a cluster.
	ListRedisFailovers(namespace string, opts metav1.ListOptions) (*redisfailoverv1.RedisFailoverList, error)
	// WatchRedisFailovers watches the redisfailovers on a cluster.
	WatchRedisFailovers(namespace string, opts metav1.ListOptions) (watch.Interface, error)
}

// RedisFailoverService is the RedisFailover service implementation using API calls to kubernetes.
type RedisFailoverService struct {
	crdClient redisfailoverclientset.Interface
	logger    log.Logger
}

// NewRedisFailoverService returns a new Workspace KubeService.
func NewRedisFailoverService(crdcli redisfailoverclientset.Interface, logger log.Logger) *RedisFailoverService {
	logger = logger.With("service", "k8s.redisfailover")
	return &RedisFailoverService{
		crdClient: crdcli,
		logger:    logger,
	}
}

// ListRedisFailovers satisfies redisfailover.Service interface.
func (r *RedisFailoverService) ListRedisFailovers(namespace string, opts metav1.ListOptions) (*redisfailoverv1.RedisFailoverList, error) {
	return r.crdClient.DatabasesV1().RedisFailovers(namespace).List(opts)
}

// WatchRedisFailovers satisfies redisfailover.Service interface.
func (r *RedisFailoverService) WatchRedisFailovers(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return r.crdClient.DatabasesV1().RedisFailovers(namespace).Watch(opts)
}
