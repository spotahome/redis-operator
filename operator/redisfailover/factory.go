package redisfailover

import (
	"time"

	kmetrics "github.com/spotahome/kooper/monitoring/metrics"
	"github.com/spotahome/kooper/operator"
	"github.com/spotahome/kooper/operator/controller"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	rfservice "github.com/spotahome/redis-operator/operator/redisfailover/service"
	"github.com/spotahome/redis-operator/service/k8s"
	"github.com/spotahome/redis-operator/service/redis"
)

const (
	// TODO: Option
	resync       = 30 * time.Second
	operatorName = "redisfailover"
)

// New will create an operator that is responsible of managing all the required stuff
// to create redis failovers.
func New(cfg Config, k8sService k8s.Services, redisClient redis.Client, mClient metrics.Instrumenter, kooperMetricsRecorder kmetrics.Recorder, logger log.Logger) operator.Operator {
	logger = logger.With("operator", operatorName)

	// Create our CRDs.
	watchedCRD := newRedisFailoverCRD(k8sService, logger)

	// Create internal services.
	rfService := rfservice.NewRedisFailoverKubeClient(k8sService, logger)
	rfChecker := rfservice.NewRedisFailoverChecker(k8sService, redisClient, logger)
	rfHealer := rfservice.NewRedisFailoverHealer(k8sService, redisClient, logger)

	// Create the handlers.
	rfHandler := NewRedisFailoverHandler(cfg, rfService, rfChecker, rfHealer, k8sService, mClient, logger)

	// Create our controller.
	ctrl := controller.NewSequential(resync, rfHandler, watchedCRD, kooperMetricsRecorder, logger.WithField("controller", "redisfailover"))

	// haproxy
	haproxyWatch := k8s.NewEndpointsRetrieve(namespace, k8sService, logger)
	haproxyHandler := (k8sService, redisClient, logger)
	haproxy := controller.NewSequential(resync, haproxyHandler, haproxyWatch, kooperMetricsRecorder, logger.WithField("controller", "redisfailover"))

	// Assemble all in an operator.
	return operator.NewOperator(watchedCRD, []controller.Controller{ctrl, haproxy}, logger)

	//return operator.NewMultiOperator([]resource.CRD{watchedCRD}, []controller.Controller{ctrl, haproxy}, logger)
}
