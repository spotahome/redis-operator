package k8s

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	redisfailoverclientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned"
	"github.com/spotahome/redis-operator/log"
)

// RedisFailover the RF service that knows how to interact with k8s to get them
type RedisFailover interface {
	// ListRedisFailovers lists the redisfailovers on a cluster.
	ListRedisFailovers(ctx context.Context, namespace string, opts metav1.ListOptions) (*redisfailoverv1.RedisFailoverList, error)
	// WatchRedisFailovers watches the redisfailovers on a cluster.
	WatchRedisFailovers(ctx context.Context, namespace string, opts metav1.ListOptions) (watch.Interface, error)
	UpdateRedisRestartedAt(namespace string, name string, restartedAt *time.Time) error
	UpdateSentinelRestartedAt(namespace string, name string, restartedAt *time.Time) error
}

// RedisFailoverService is the RedisFailover service implementation using API calls to kubernetes.
type RedisFailoverService struct {
	k8sCli redisfailoverclientset.Interface
	logger log.Logger
}

const (
	redisRestartedAtPatch = `{
		"status": {
			"redisRestartedAt": "%s"
		}
	}`
	sentinelRestartedAtPatch = `{
		"status": {
			"sentinelRestartedAt": "%s"
		}
	}`
)

// NewRedisFailoverService returns a new Workspace KubeService.
func NewRedisFailoverService(k8scli redisfailoverclientset.Interface, logger log.Logger) *RedisFailoverService {
	logger = logger.With("service", "k8s.redisfailover")
	return &RedisFailoverService{
		k8sCli: k8scli,
		logger: logger,
	}
}

// ListRedisFailovers satisfies redisfailover.Service interface.
func (r *RedisFailoverService) ListRedisFailovers(ctx context.Context, namespace string, opts metav1.ListOptions) (*redisfailoverv1.RedisFailoverList, error) {
	return r.k8sCli.DatabasesV1().RedisFailovers(namespace).List(ctx, opts)
}

// WatchRedisFailovers satisfies redisfailover.Service interface.
func (r *RedisFailoverService) WatchRedisFailovers(ctx context.Context, namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return r.k8sCli.DatabasesV1().RedisFailovers(namespace).Watch(ctx, opts)
}

// UpdateRedisRestartedAt updates redis restartedAt status
func (r *RedisFailoverService) UpdateRedisRestartedAt(namespace string, name string, restartedAt *time.Time) error {
	ctx := context.TODO()
	redisfailoverIf := r.k8sCli.DatabasesV1().RedisFailovers(namespace)
	if restartedAt == nil {
		t := time.Now().UTC()
		restartedAt = &t
	}
	patch := fmt.Sprintf(redisRestartedAtPatch, restartedAt.Format(time.RFC3339))
	_, err := redisfailoverIf.Patch(ctx, name, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}

// UpdateSentinelRestartedAt updates redis restartedAt status
func (r *RedisFailoverService) UpdateSentinelRestartedAt(namespace string, name string, restartedAt *time.Time) error {
	ctx := context.TODO()
	redisfailoverIf := r.k8sCli.DatabasesV1().RedisFailovers(namespace)
	if restartedAt == nil {
		t := time.Now().UTC()
		restartedAt = &t
	}
	patch := fmt.Sprintf(sentinelRestartedAtPatch, restartedAt.Format(time.RFC3339))
	_, err := redisfailoverIf.Patch(ctx, name, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}
