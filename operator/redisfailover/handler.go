package redisfailover

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	rfservice "github.com/spotahome/redis-operator/operator/redisfailover/service"
	"github.com/spotahome/redis-operator/operator/redisfailover/util"
	"github.com/spotahome/redis-operator/service/k8s"
)

const (
	rfLabelManagedByKey = "app.kubernetes.io/managed-by"
	rfLabelNameKey      = "redisfailovers.databases.spotahome.com/name"
)

var (
	defaultLabels = map[string]string{
		rfLabelManagedByKey: operatorName,
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
}

// NewRedisFailoverHandler returns a new RF handler
func NewRedisFailoverHandler(config Config, rfService rfservice.RedisFailoverClient, rfChecker rfservice.RedisFailoverCheck, rfHealer rfservice.RedisFailoverHeal, k8sservice k8s.Service, mClient metrics.Instrumenter, logger log.Logger) *RedisFailoverHandler {
	return &RedisFailoverHandler{
		config:     config,
		rfService:  rfService,
		rfChecker:  rfChecker,
		rfHealer:   rfHealer,
		mClient:    mClient,
		k8sservice: k8sservice,
		logger:     logger,
	}
}

// Add will ensure the redis failover is in the expected state.
func (r *RedisFailoverHandler) Add(_ context.Context, obj runtime.Object) error {
	rf, ok := obj.(*redisfailoverv1.RedisFailover)
	if !ok {
		return fmt.Errorf("can't handle the received object: not a redisfailover")
	}

	if err := rf.Validate(); err != nil {
		r.mClient.SetClusterError(rf.Namespace, rf.Name)
		return err
	}

	// Create owner refs so the objects manager by this handler have ownership to the
	// received RF.
	oRefs := r.createOwnerReferences(rf)

	// Create the labels every object derived from this need to have.
	labels := r.getLabels(rf)

	if err := r.Ensure(rf, labels, oRefs); err != nil {
		r.mClient.SetClusterError(rf.Namespace, rf.Name)
		return err
	}

	if err := r.CheckAndHeal(rf); err != nil {
		r.mClient.SetClusterError(rf.Namespace, rf.Name)
		return err
	}

	r.mClient.SetClusterOK(rf.Namespace, rf.Name)
	return nil
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

// getLabels merges the labels (dynamic and operator static ones).
func (r *RedisFailoverHandler) getLabels(rf *redisfailoverv1.RedisFailover) map[string]string {
	dynLabels := map[string]string{
		rfLabelNameKey: rf.Name,
	}

	// Filter the labels based on the whitelist
	filteredCustomLabels := make(map[string]string)
	if rf.Spec.LabelWhitelist != nil && len(rf.Spec.LabelWhitelist) != 0 {
		for _, regex := range rf.Spec.LabelWhitelist {
			compiledRegexp, err := regexp.Compile(regex)
			if err != nil {
				r.logger.Errorf("Unable to compile label whitelist regex '%s', ignoring it.", regex)
				continue
			}
			for labelKey, labelValue := range rf.Labels {
				if match := compiledRegexp.MatchString(labelKey); match {
					filteredCustomLabels[labelKey] = labelValue
				}
			}
		}
	} else {
		// If no whitelist is specified then don't filter the labels.
		filteredCustomLabels = rf.Labels
	}
	return util.MergeLabels(defaultLabels, dynLabels, filteredCustomLabels)
}

func (w *RedisFailoverHandler) createOwnerReferences(rf *redisfailoverv1.RedisFailover) []metav1.OwnerReference {
	rfvk := redisfailoverv1.VersionKind(redisfailoverv1.RFKind)
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(rf, rfvk),
	}
}
