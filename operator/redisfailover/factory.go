package redisfailover

import (
	"context"
	"regexp"
	"time"

	"github.com/spotahome/kooper/v2/controller"
	"github.com/spotahome/kooper/v2/controller/leaderelection"
	kooperlog "github.com/spotahome/kooper/v2/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	rfservice "github.com/spotahome/redis-operator/operator/redisfailover/service"
	"github.com/spotahome/redis-operator/service/k8s"
	"github.com/spotahome/redis-operator/service/redis"
)

const (
	resync       = 30 * time.Second
	operatorName = "redis-operator"
	lockKey      = "redis-failover-lease"
)

// New will create an operator that is responsible of managing all the required stuff
// to create redis failovers.
func New(cfg Config, k8sService k8s.Services, k8sClient kubernetes.Interface, lockNamespace string, redisClient redis.Client, kooperMetricsRecorder metrics.Recorder, logger log.Logger) (controller.Controller, error) {
	// Create internal services.
	rfService := rfservice.NewRedisFailoverKubeClient(k8sService, logger, kooperMetricsRecorder)
	rfChecker := rfservice.NewRedisFailoverChecker(k8sService, redisClient, logger, kooperMetricsRecorder)
	rfHealer := rfservice.NewRedisFailoverHealer(k8sService, redisClient, logger)

	// Create the handlers.
	rfHandler := NewRedisFailoverHandler(cfg, rfService, rfChecker, rfHealer, k8sService, kooperMetricsRecorder, logger)
	rfRetriever := NewRedisFailoverRetriever(cfg, k8sService)

	kooperLogger := kooperlogger{Logger: logger.WithField("operator", "redisfailover")}
	// Leader election service.
	leSVC, err := leaderelection.NewDefault(lockKey, lockNamespace, k8sClient, kooperLogger)
	if err != nil {
		return nil, err
	}

	// Create our controller.
	return controller.New(&controller.Config{
		Handler:           rfHandler,
		Retriever:         rfRetriever,
		LeaderElector:     leSVC,
		MetricsRecorder:   kooperMetricsRecorder,
		Logger:            kooperLogger,
		Name:              "redisfailover",
		ResyncInterval:    resync,
		ConcurrentWorkers: cfg.Concurrency,
	})
}

func NewRedisFailoverRetriever(cfg Config, cli k8s.Services) controller.Retriever {
	isNamespaceSupported := func(rf redisfailoverv1.RedisFailover) bool {
		match, _ := regexp.Match(cfg.SupportedNamespacesRegex, []byte(rf.Namespace))
		return match
	}
	// check in the startup whether the regex compiles

	return controller.MustRetrieverFromListerWatcher(&cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			rfList, err := cli.ListRedisFailovers(context.Background(), "", options)
			if err != nil {
				return rfList, err
			}

			targetRFList := make([]redisfailoverv1.RedisFailover, 0)
			for _, rf := range rfList.Items {
				if isNamespaceSupported(rf) {
					targetRFList = append(targetRFList, rf)
				}
			}
			rfList.Items = targetRFList

			return rfList, err
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			watcher, err := cli.WatchRedisFailovers(context.Background(), "", options)
			watcher = watch.Filter(watcher, func(event watch.Event) (watch.Event, bool) {
				rf, ok := event.Object.(*redisfailoverv1.RedisFailover)
				if !ok {
					return event, false
				}
				return event, isNamespaceSupported(*rf)
			})
			return watcher, err
		},
	})
}

type kooperlogger struct {
	log.Logger
}

func (k kooperlogger) WithKV(kv kooperlog.KV) kooperlog.Logger {
	return kooperlogger{Logger: k.Logger.WithFields(kv)}
}
