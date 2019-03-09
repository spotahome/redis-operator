package redisfailover

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	redisfailoverv1alpha2 "github.com/spotahome/redis-operator/api/redisfailover/v1alpha2"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	rfservice "github.com/spotahome/redis-operator/operator/redisfailover/service"
	"github.com/spotahome/redis-operator/operator/redisfailover/util"
	"github.com/spotahome/redis-operator/service/k8s"
)

const (
	redisFailoverLabelKey = "redisfailover"
)

var (
	defaultLabels = map[string]string{
		"creator": operatorName,
	}
)

// RedisFailoverHandler is the Redis Failover handler. This handler will create the required
// resources that a RF needs.
type RedisFailoverHandler struct {
	config     Config
	k8sservice k8s.Service
	rfService  rfservice.RedisFailoverClient
	rfChecker  rfservice.RedisFailoverCheck
	rfHealer   rfservice.RedisFailoverHeal
	mClient    metrics.Instrumenter
	logger     log.Logger
	labels     map[string]string
}

// NewRedisFailoverHandler returns a new RF handler
func NewRedisFailoverHandler(config Config, rfService rfservice.RedisFailoverClient, rfChecker rfservice.RedisFailoverCheck, rfHealer rfservice.RedisFailoverHeal, k8sservice k8s.Service, mClient metrics.Instrumenter, logger log.Logger) *RedisFailoverHandler {
	// Set non dynamic operator labels(the ones that every resource created by the operator will have).
	labels := util.MergeLabels(config.Labels, defaultLabels)

	return &RedisFailoverHandler{
		config:     config,
		rfService:  rfService,
		rfChecker:  rfChecker,
		rfHealer:   rfHealer,
		mClient:    mClient,
		k8sservice: k8sservice,
		logger:     logger,
		labels:     labels,
	}
}

// Add will ensure the redis failover is in the expected state.
func (r *RedisFailoverHandler) Add(_ context.Context, obj runtime.Object) error {
	rf, ok := obj.(*redisfailoverv1alpha2.RedisFailover)
	if !ok {
		return fmt.Errorf("can't handle redis failover state, parentLabels map[string]string, ownerRefs []metav1.OwnerReferencenot a redisfailover object")
	}

	if err := rf.Validate(); err != nil {
		r.mClient.SetClusterError(rf.Namespace, rf.Name)
		return err
	}

	// Create owner refs so the objects manager by this handler have ownership to the
	// received RF.
	oRefs := r.createOwnerReferences(rf)

	// Create the labels every object derived from this need to have.
	labels := r.mergeLabels(rf)

	if err := r.Ensure(rf, labels, oRefs); err != nil {
		return err
	}

	return r.CheckAndHeal(rf, oRefs)
}

// Delete handles the deletion of a RF.
func (r *RedisFailoverHandler) Delete(_ context.Context, name string) error {
	n := strings.Split(name, "/")
	if len(n) >= 2 {
		r.mClient.DeleteCluster(n[0], n[1])
	}
	// No need to do anything, it will be handled by the owner reference done
	// on the creation.
	r.logger.Debugf("ignoring, kubernetes GCs all using the objects OwnerReference metadata")
	return nil
}

// mergeLabels merges all the labels (dynamic and operator static ones).
func (r *RedisFailoverHandler) mergeLabels(rf *redisfailoverv1alpha2.RedisFailover) map[string]string {
	dynLabels := map[string]string{
		redisFailoverLabelKey: rf.Name,
	}
	return util.MergeLabels(r.labels, dynLabels, rf.Labels)
}

func (w *RedisFailoverHandler) createOwnerReferences(rf *redisfailoverv1alpha2.RedisFailover) []metav1.OwnerReference {
	rfvk := redisfailoverv1alpha2.VersionKind(redisfailoverv1alpha2.RFKind)
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(rf, rfvk),
	}
}
