package redisfailover

import (
	redisfailoverv1alpha2 "github.com/spotahome/redis-operator/api/redisfailover/v1alpha2"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/service/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// redisfailoverCRD is the crd redis failover
type redisfailoverCRD struct {
	service k8s.Services
	logger  log.Logger
}

func newRedisFailoverCRD(service k8s.Services, logger log.Logger) *redisfailoverCRD {
	logger = logger.With("crd", "redisfailover")
	return &redisfailoverCRD{
		service: service,
		logger:  logger,
	}
}

// Initialize satisfies resource.crd interface.
func (w *redisfailoverCRD) Initialize() error {
	crd := k8s.CRDConf{
		Kind:       redisfailoverv1alpha2.RFKind,
		NamePlural: redisfailoverv1alpha2.RFNamePlural,
		Group:      redisfailoverv1alpha2.SchemeGroupVersion.Group,
		Version:    redisfailoverv1alpha2.SchemeGroupVersion.Version,
		Scope:      redisfailoverv1alpha2.RFScope,
	}

	return w.service.EnsureCRD(crd)
}

// GetListerWatcher satisfies resource.crd interface (and retrieve.Retriever).
func (w *redisfailoverCRD) GetListerWatcher() cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return w.service.ListRedisFailovers("", options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return w.service.WatchRedisFailovers("", options)
		},
	}
}

// GetObject satisfies resource.crd interface (and retrieve.Retriever).
func (w *redisfailoverCRD) GetObject() runtime.Object {
	return &redisfailoverv1alpha2.RedisFailover{}
}
