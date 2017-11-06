package failover

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/spotahome/redis-operator/pkg/clock"
	"github.com/spotahome/redis-operator/pkg/config"
	"github.com/spotahome/redis-operator/pkg/crd"
	"github.com/spotahome/redis-operator/pkg/log"
	"github.com/spotahome/redis-operator/pkg/metrics"
	"github.com/spotahome/redis-operator/pkg/redis"
)

// NewRedisFailoverCRD returns a redis failover CRD
func NewRedisFailoverCRD(metricsClient metrics.Instrumenter, clientset kubernetes.Interface, restConfig rest.Config, signalChan chan int, maxThreads int) (*crd.CRD, error) {
	logger := log.Base()
	// Create the Redis failover client that will talk with K8s
	client := NewRedisFailoverKubeClient(clientset, clock.Base(), logger)

	// Create our controller that will handle the K8s events
	rfChecker := NewRedisFailoverChecker(metricsClient, client, redis.New(), clock.Base(), logger)
	eventHandler := NewRedisFailoverControllerAsync(metricsClient, client, logger, &RedisFailoverTransformer{}, rfChecker, maxThreads)

	// create the CRD with all the required pieces
	crd, err := crd.NewCRD(restConfig, signalChan, config.Kind, config.Domain, config.Version, config.APIName, &RedisFailover{}, &RedisFailoverList{}, &eventHandler)
	if err != nil {
		return nil, err
	}

	return crd, nil
}
