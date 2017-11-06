package failover

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spotahome/redis-operator/pkg/config"
	"github.com/spotahome/redis-operator/pkg/log"
	"github.com/spotahome/redis-operator/pkg/metrics"
)

const (
	logNameField      = "redisfailover"
	logNamespaceField = "namespace"
)

// RedisFailoverController handles all the events related with a RedisFailover TPR
type RedisFailoverController struct {
	Client      RedisFailoverClient
	logger      log.Logger
	transformer Transformer
	checker     RedisFailoverCheck
	metrics     metrics.Instrumenter
}

// NewRedisFailoverController creates a RedisFailoverController
func NewRedisFailoverController(metricsClient metrics.Instrumenter, client RedisFailoverClient, logger log.Logger, transformer Transformer, checker RedisFailoverCheck) RedisFailoverController {
	return RedisFailoverController{
		Client:      client,
		logger:      logger,
		transformer: transformer,
		checker:     checker,
		metrics:     metricsClient,
	}
}

// OnStatus satisfies EventHandler interface.
// It checks if the sentinels config are ok, and if not, tries to fix it
func (r *RedisFailoverController) OnStatus() {
	rfs, err := r.Client.GetAllRedisfailovers()
	if err != nil {
		r.logger.Error(err)
		return
	}

	var creating, running, failed float64

	for _, rf := range rfs.Items {
		c, r, f := r.OnRFStatus(rf)
		creating += c
		running += r
		failed += f
	}

	// Set metrics.
	r.metrics.SetClustersCreating(creating)
	r.metrics.SetClustersRunning(running)
	r.metrics.SetClustersFailed(failed)
}

// OnRFStatus checks ONE redis failover.
func (r *RedisFailoverController) OnRFStatus(rf RedisFailover) (float64, float64, float64) {
	var creating, running, failed float64
	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)
	r.validateSpec(&rf)
	if rf.Status.Phase != PhaseRunning || rf.Status.Conditions[len(rf.Status.Conditions)-1].Type != ConditionReady {
		creating++
		logger.Debugf("RedisFailover %s in namespace %s is not Ready, skipping", rf.Metadata.Name, rf.Metadata.Namespace)
		return creating, running, failed
	}
	if err := r.checker.Check(&rf); err != nil {
		failed++
		logger.Errorf("RedisFailover %s in namespace %s has the following error: %s", rf.Metadata.Name, rf.Metadata.Namespace, err)
	} else {
		running++
		logger.Debugf("RedisFailover %s in namespace %s is ok", rf.Metadata.Name, rf.Metadata.Namespace)
		if master, err := r.checker.GetMaster(&rf); err == nil && rf.Status.Master != master {
			rf.Status.SetMaster(master)
			r.Client.UpdateStatus(&rf)
		}
	}
	return creating, running, failed
}

// OnAdd satisfies EventHandler interface.
func (r *RedisFailoverController) OnAdd(obj interface{}) {
	x := obj.(*RedisFailover)
	var err error
	logger := r.logger.WithField(logNameField, x.Metadata.Name).WithField(logNamespaceField, x.Metadata.Namespace)
	r.metrics.IncAddEventHandled(x.Metadata.Name)

	if err = r.validateSpec(x); err != nil {
		logger.Error(err.Error())
		return
	}

	// If PhaseNone this RedisFailover is a new one
	if x.Status.Phase != PhaseNone {
		if x.Status.Phase == PhaseRunning {
			redis, err2 := r.Client.GetRedisStatefulset(x)
			if err2 != nil {
				return
			}
			sentinel, err2 := r.Client.GetSentinelDeployment(x)
			if err2 != nil {
				return
			}
			rSettings, err2 := r.transformer.StatefulsetToRedisSettings(redis)
			if err2 != nil {
				return
			}
			sSettings, err2 := r.transformer.DeploymentToSentinelSettings(sentinel)
			if err2 != nil {
				return
			}
			existingRF := &RedisFailover{
				Metadata: metav1.ObjectMeta{
					Name:      x.Metadata.Name,
					Namespace: x.Metadata.Namespace,
				},
				Spec: RedisFailoverSpec{
					Redis:    *rSettings,
					Sentinel: *sSettings,
				},
			}
			logger.Debug("Redis Failover already created. Checking if it is updated")
			r.OnUpdate(existingRF, x)
			return
		}
	}

	if x.Status.Phase == PhaseNone {
		logger.Info("Deploying Redis failover...")
		x.Status.SetPhase(PhaseCreating)
		x.Status.SetNotReadyCondition()
		if _, err = r.Client.UpdateStatus(x); err != nil {
			logger.Errorf("Error updating status: %s", err)
			return
		}
	} else {
		logger.Info("Resuming deployment of Redis failover...")
	}

	if _, err = r.Client.GetBootstrapPod(x); err != nil {
		if _, err = r.Client.GetSentinelDeployment(x); err != nil {
			if _, err = r.Client.GetRedisStatefulset(x); err != nil {
				logger.Debug("Creating bootstrap pod...")
				if err = r.Client.CreateBootstrapPod(x); err != nil {
					logger.Errorf("Could not create bootstrap pod: %s", err)
					x.Status.SetPhase(PhaseFailed)
					r.Client.UpdateStatus(x)
					return
				}
				logger.Debug("Pod created!")
			}
		}
	}

	if _, err = r.Client.GetSentinelService(x); err != nil {
		logger.Debug("Creating Sentinel service...")
		if err = r.Client.CreateSentinelService(x); err != nil {
			logger.Errorf("Could not create sentinel service: %s", err)
			x.Status.SetPhase(PhaseFailed)
			r.Client.UpdateStatus(x)
			return
		}
		logger.Debug("Service created!")
	}

	if _, err = r.Client.GetSentinelDeployment(x); err != nil {
		logger.Debug("Creating Sentinel deployment...")
		if err = r.Client.CreateSentinelDeployment(x); err != nil {
			logger.Errorf("Could not create sentinel deployment: %s", err)
			x.Status.SetPhase(PhaseFailed)
			r.Client.UpdateStatus(x)
			return
		}
		logger.Debug("Deployment created!")
	}

	if x.Spec.Redis.Exporter {
		if _, err = r.Client.GetRedisService(x); err != nil {
			logger.Debug("Creating redis service...")
			if err = r.Client.CreateRedisService(x); err != nil {
				logger.Errorf("Could not create redis service: %s", err)
				x.Status.SetPhase(PhaseFailed)
				r.Client.UpdateStatus(x)
				return
			}
			logger.Debug("Service created!")
		}
	}

	if _, err = r.Client.GetRedisStatefulset(x); err != nil {
		logger.Debug("Creating Redis statefulset...")
		if err = r.Client.CreateRedisStatefulset(x); err != nil {
			logger.Errorf("Could not create Redis statefulset: %s", err)
			x.Status.SetPhase(PhaseFailed)
			r.Client.UpdateStatus(x)
			return
		}
		logger.Debug("Statefulset created!")
	}

	if _, err = r.Client.GetBootstrapPod(x); err == nil {
		logger.Debug("Deleting bootstrap pod...")
		if err = r.Client.DeleteBootstrapPod(x); err != nil {
			logger.Errorf("Could not delete Bootstrap pod: %s", err)
			x.Status.SetPhase(PhaseFailed)
			r.Client.UpdateStatus(x)
			return
		}
		logger.Debug("Pod deleted!")
	}

	x.Status.SetPhase(PhaseRunning)
	x.Status.SetReadyCondition()
	if _, err = r.Client.UpdateStatus(x); err != nil {
		logger.Errorf("Error updating status: %s", err)
		return
	}
	logger.Info("Redis failover deployed!")
}

// OnUpdate satisfies EventHandler interface.
func (r *RedisFailoverController) OnUpdate(oldObj, newObj interface{}) {
	old := oldObj.(*RedisFailover)
	new := newObj.(*RedisFailover)
	logger := r.logger.WithField(logNameField, old.Metadata.Name).WithField(logNamespaceField, old.Metadata.Namespace)
	r.metrics.IncUpdateEventHandled(old.Metadata.Name)

	r.validateSpec(new)
	r.validateSpec(old)

	new.GetQuorum()
	old.GetQuorum()

	if new.Spec == old.Spec {
		return
	}

	if new.Spec.Sentinel != old.Spec.Sentinel {
		logger.Info("Updating Sentinel deployment...")
		logger.Debugf("OLD: \n%+v", old.Spec.Sentinel)
		logger.Debugf("NEW: \n%+v", new.Spec.Sentinel)
		if old.Spec.Sentinel.Replicas < new.Spec.Sentinel.Replicas {
			new.Status.AppendScalingSentinelUpCondition(old.Spec.Sentinel.Replicas, new.Spec.Sentinel.Replicas)
		} else if old.Spec.Sentinel.Replicas > new.Spec.Sentinel.Replicas {
			new.Status.AppendScalingSentinelDownCondition(old.Spec.Sentinel.Replicas, new.Spec.Sentinel.Replicas)
		} else {
			new.Status.AppendUpdatingSentinelCondition("Change the resources/limits")
		}
		if _, err2 := r.Client.UpdateStatus(new); err2 != nil {
			logger.Errorf("Error updating status: %s", err2)
		}
		if err2 := r.Client.UpdateSentinelDeployment(new); err2 != nil {
			logger.Errorf("Error updating Sentinel deployment: %s", err2)
			return
		}
		logger.Info("Sentinel Deployment updated!")
	}
	if new.Spec.Redis != old.Spec.Redis {
		logger.Info("Updating Redis statefulset...")
		logger.Debugf("OLD: \n%+v", old.Spec.Redis)
		logger.Debugf("NEW: \n%+v", new.Spec.Redis)
		if old.Spec.Redis.Replicas < new.Spec.Redis.Replicas {
			new.Status.AppendScalingRedisUpCondition(old.Spec.Redis.Replicas, new.Spec.Redis.Replicas)
		} else if old.Spec.Redis.Replicas > new.Spec.Redis.Replicas {
			new.Status.AppendScalingRedisDownCondition(old.Spec.Redis.Replicas, new.Spec.Redis.Replicas)
		} else {
			new.Status.AppendUpdatingRedisCondition("Change the resources/limits")
		}
		if _, err2 := r.Client.UpdateStatus(new); err2 != nil {
			logger.Errorf("Error updating status: %s", err2)
		}
		if err2 := r.Client.UpdateRedisStatefulset(new); err2 != nil {
			logger.Errorf("Error updating Redis statefulset: %s", err2)
			return
		}
		logger.Info("Redis statefulset updated!")
	}
	new.Status.SetReadyCondition()
	if _, err := r.Client.UpdateStatus(new); err != nil {
		logger.Errorf("Error updating status: %s", err)
	}
}

// OnDelete satisfies EventHandler interface.
func (r *RedisFailoverController) OnDelete(obj interface{}) {
	x := obj.(*RedisFailover)

	logger := r.logger.WithField(logNameField, x.Metadata.Name).WithField(logNamespaceField, x.Metadata.Namespace)
	r.metrics.IncDeleteEventHandled(x.Metadata.Name)

	logger.Info("Deleting Redis Failover...")

	logger.Debug("Deleting redis service...")
	r.Client.DeleteRedisService(x)
	logger.Debug("Service deleted!")

	logger.Debug("Deleting redis statefulset...")
	r.Client.DeleteRedisStatefulset(x)
	logger.Debug("Statefulset deleted!")

	logger.Debug("Deleting sentinel service...")
	r.Client.DeleteSentinelService(x)
	logger.Debug("Service deleted!")

	logger.Debug("Deleting sentinel deployment...")
	r.Client.DeleteSentinelDeployment(x)
	logger.Debug("Deployment deleted!")

	logger.Info("Redis Failover deleted!")
}

func (r *RedisFailoverController) validateSpec(x *RedisFailover) error {
	if x.Spec.Redis.Replicas == 0 {
		x.Spec.Redis.Replicas = 3
	}

	if x.Spec.Sentinel.Replicas == 0 {
		x.Spec.Sentinel.Replicas = 3
	}

	if x.Spec.Redis.Replicas < 3 {
		err := fmt.Errorf("Number of redis replicas not valid. Got: %d", x.Spec.Redis.Replicas)
		return err
	}

	if x.Spec.Sentinel.Replicas < 3 {
		err := fmt.Errorf("Number of sentinel replicas not valid. Got %d", x.Spec.Sentinel.Replicas)
		return err
	}

	if x.Spec.Redis.Version == "" {
		x.Spec.Redis.Version = config.RedisImageVersion
	}

	return nil
}

// RedisFailoverControllerAsync handles all the events related with a RedisFailover TPR but asyncronous
type RedisFailoverControllerAsync struct {
	rfc          RedisFailoverController
	locks        map[string]*sync.Mutex
	changingLock *sync.Mutex
	changing     bool
	semaphore    *semaphore.Weighted
}

const (
	semaphoreAquireNumber = int64(1)
	semaphoreWaitTime     = 1 * time.Second
)

// NewRedisFailoverControllerAsync creates a RedisFailoverControllerAsync
func NewRedisFailoverControllerAsync(metricsClient metrics.Instrumenter, client RedisFailoverClient, logger log.Logger, transformer Transformer, checker RedisFailoverCheck, maxThreads int) RedisFailoverControllerAsync {
	return RedisFailoverControllerAsync{
		rfc: RedisFailoverController{
			Client:      client,
			logger:      logger,
			transformer: transformer,
			checker:     checker,
			metrics:     metricsClient,
		},
		locks:     map[string]*sync.Mutex{},
		semaphore: semaphore.NewWeighted(int64(maxThreads)),
	}
}

func (r *RedisFailoverControllerAsync) getIdentifier(obj interface{}) string {
	x := obj.(*RedisFailover)
	return fmt.Sprintf("%s-%s", x.Metadata.Namespace, x.Metadata.Name)
}

func (r *RedisFailoverControllerAsync) getMutex(obj interface{}) *sync.Mutex {
	name := r.getIdentifier(obj)
	r.rfc.logger.Debugf("Getting mutex for %s from mutex array", name)
	mu, ok := r.locks[name]
	if !ok {
		r.rfc.logger.Debugf("Lock for %s doesn't exists, creating...", name)
		mu = &sync.Mutex{}
		r.locks[name] = mu
	}
	return mu
}

func (r *RedisFailoverControllerAsync) getSemaphore() {
	for !r.semaphore.TryAcquire(semaphoreAquireNumber) {
		r.rfc.logger.Debug("Maximum threads reached, waiting...")
		time.Sleep(semaphoreWaitTime)
	}
}

func (r *RedisFailoverControllerAsync) releaseSemaphore() {
	r.semaphore.Release(semaphoreAquireNumber)
}

// OnAdd satisfies EventHandler interface.
func (r *RedisFailoverControllerAsync) OnAdd(obj interface{}) {
	r.getSemaphore()
	go func() {
		mu := r.getMutex(obj)
		mu.Lock()
		defer r.releaseSemaphore()
		defer mu.Unlock()
		r.rfc.OnAdd(obj)
	}()
}

// OnUpdate satisfies EventHandler interface.
func (r *RedisFailoverControllerAsync) OnUpdate(oldObj, newObj interface{}) {
	r.getSemaphore()
	go func() {
		mu := r.getMutex(oldObj)
		mu.Lock()
		defer r.releaseSemaphore()
		defer mu.Unlock()
		r.rfc.OnUpdate(oldObj, newObj)
	}()
}

// OnDelete satisfies EventHandler interface.
func (r *RedisFailoverControllerAsync) OnDelete(obj interface{}) {
	r.getSemaphore()
	go func() {
		mu := r.getMutex(obj)
		mu.Lock()
		defer r.releaseSemaphore()
		defer mu.Unlock()
		r.rfc.OnDelete(obj)
	}()
}

// OnStatus satisfies EventHandler interface.
func (r *RedisFailoverControllerAsync) OnStatus() {
	rfs, err := r.rfc.Client.GetAllRedisfailovers()
	if err != nil {
		r.rfc.logger.Error(err)
		return
	}
	for _, rf := range rfs.Items {
		rf := rf
		r.getSemaphore()
		go func() {
			mu := r.getMutex(&rf)
			mu.Lock()
			defer r.releaseSemaphore()
			defer mu.Unlock()
			creating, running, failed := r.rfc.OnRFStatus(rf)
			r.rfc.metrics.SetClustersCreating(creating)
			r.rfc.metrics.SetClustersRunning(running)
			r.rfc.metrics.SetClustersFailed(failed)
		}()
	}
}
