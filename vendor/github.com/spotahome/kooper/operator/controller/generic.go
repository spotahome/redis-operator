package controller

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/monitoring/metrics"
	"github.com/spotahome/kooper/operator/controller/leaderelection"
	"github.com/spotahome/kooper/operator/handler"
	"github.com/spotahome/kooper/operator/retrieve"
)

// Span tag and log keys.
const (
	kubernetesObjectKeyKey         = "kubernetes.object.key"
	kubernetesObjectNSKey          = "kubernetes.object.namespace"
	kubernetesObjectNameKey        = "kubernetes.object.name"
	eventKey                       = "event"
	kooperControllerKey            = "kooper.controller"
	processedTimesKey              = "kubernetes.object.total_processed_times"
	retriesRemainingKey            = "kubernetes.object.retries_remaining"
	processingRetryKey             = "kubernetes.object.processing_retry"
	retriesExecutedKey             = "kubernetes.object.retries_consumed"
	controllerNameKey              = "controller.cfg.name"
	controllerResyncKey            = "controller.cfg.resync_interval"
	controllerMaxRetriesKey        = "controller.cfg.max_retries"
	controllerConcurrentWorkersKey = "controller.cfg.concurrent_workers"
	successKey                     = "success"
	messageKey                     = "message"
)

// generic controller is a controller that can be used to create different kind of controllers.
type generic struct {
	queue       workqueue.RateLimitingInterface // queue will have the jobs that the controller will get and send to handlers.
	informer    cache.SharedIndexInformer       // informer will notify be inform us about resource changes.
	handler     handler.Handler                 // handler is where the logic of resource processing.
	handlerName string                          // handlerName will be used to identify and give more insight about metrics.
	running     bool
	runningMu   sync.Mutex
	cfg         Config
	tracer      opentracing.Tracer // use directly opentracing API because it's not an implementation.
	metrics     metrics.Recorder
	leRunner    leaderelection.Runner
	logger      log.Logger
}

// NewSequential creates a new controller that will process the received events sequentially.
// This constructor is just a wrapper to help bootstrapping default sequential controller.
func NewSequential(resync time.Duration, handler handler.Handler, retriever retrieve.Retriever, metricRecorder metrics.Recorder, logger log.Logger) Controller {
	cfg := &Config{
		ConcurrentWorkers: 1,
		ResyncInterval:    resync,
	}
	return New(cfg, handler, retriever, nil, nil, metricRecorder, logger)
}

// NewConcurrent creates a new controller that will process the received events concurrently.
// This constructor is just a wrapper to help bootstrapping default concurrent controller.
func NewConcurrent(concurrentWorkers int, resync time.Duration, handler handler.Handler, retriever retrieve.Retriever, metricRecorder metrics.Recorder, logger log.Logger) (Controller, error) {
	if concurrentWorkers < 2 {
		return nil, fmt.Errorf("%d is not a valid concurrency workers ammount for a concurrent controller", concurrentWorkers)
	}

	cfg := &Config{
		ConcurrentWorkers: concurrentWorkers,
		ResyncInterval:    resync,
	}
	return New(cfg, handler, retriever, nil, nil, metricRecorder, logger), nil
}

// New creates a new controller that can be configured using the cfg parameter.
func New(cfg *Config, handler handler.Handler, retriever retrieve.Retriever, leaderElector leaderelection.Runner, tracer opentracing.Tracer, metricRecorder metrics.Recorder, logger log.Logger) Controller {
	// Sets the required default configuration.
	cfg.setDefaults()

	// Default logger.
	if logger == nil {
		logger = &log.Std{}
		logger.Warningf("no logger specified, fallback to default logger, to disable logging use dummy logger")
	}

	// Default metrics recorder.
	if metricRecorder == nil {
		metricRecorder = metrics.Dummy
		logger.Warningf("no metrics recorder specified, disabling metrics")
	}

	// Default tracer.
	if tracer == nil {
		tracer = &opentracing.NoopTracer{}
	}

	// Get a handler name for the metrics based on the type of the handler.
	handlerName := reflect.TypeOf(handler).String()

	// Create the queue that will have our received job changes. It's rate limited so we don't have problems when
	// a job processing errors every time is processed in a loop.
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// store is the internal cache where objects will be store.
	store := cache.Indexers{}
	informer := cache.NewSharedIndexInformer(retriever.GetListerWatcher(), retriever.GetObject(), cfg.ResyncInterval, store)

	// Set up our informer event handler.
	// Objects are already in our local store. Add only keys/jobs on the queue so they can bre processed
	// afterwards.
	informer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
				metricRecorder.IncResourceAddEventQueued(handlerName)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
				metricRecorder.IncResourceAddEventQueued(handlerName)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
				metricRecorder.IncResourceDeleteEventQueued(handlerName)
			}
		},
	}, cfg.ResyncInterval)

	// Create our generic controller object.
	return &generic{
		queue:       queue,
		informer:    informer,
		logger:      logger,
		metrics:     metricRecorder,
		tracer:      tracer,
		handler:     handler,
		handlerName: handlerName,
		leRunner:    leaderElector,
		cfg:         *cfg,
	}
}

func (g *generic) isRunning() bool {
	g.runningMu.Lock()
	defer g.runningMu.Unlock()
	return g.running
}

func (g *generic) setRunning(running bool) {
	g.runningMu.Lock()
	defer g.runningMu.Unlock()
	g.running = running
}

// Run will run the controller.
func (g *generic) Run(stopC <-chan struct{}) error {
	// Check if leader election is required.
	if g.leRunner != nil {
		return g.leRunner.Run(func() error {
			return g.run(stopC)
		})
	}

	return g.run(stopC)
}

// run is the real run of the controller.
func (g *generic) run(stopC <-chan struct{}) error {
	if g.isRunning() {
		return fmt.Errorf("controller already running")
	}

	g.logger.Infof("starting controller")
	// Set state of controller.
	g.setRunning(true)
	defer g.setRunning(false)

	// Shutdown when Run is stopped so we can process the last items and the queue doesn't
	// accept more jobs.
	defer g.queue.ShutDown()

	// Run the informer so it starts listening to resource events.
	go g.informer.Run(stopC)

	// Wait until our store, jobs... stuff is synced (first list on resource, resources on store and jobs on queue).
	if !cache.WaitForCacheSync(stopC, g.informer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	// Start our resource processing worker, if finishes then restart the worker. The workers should
	// not end.
	for i := 0; i < g.cfg.ConcurrentWorkers; i++ {
		go func() {
			wait.Until(g.runWorker, time.Second, stopC)
		}()
	}

	// Until will be running our workers in a continuous way (and re run if they fail). But
	// when stop signal is received we must stop.
	<-stopC
	g.logger.Infof("stopping controller")

	return nil
}

// runWorker will start a processing loop on event queue.
func (g *generic) runWorker() {
	for {
		// Process newxt queue job, if needs to stop processing it will return true.
		if g.getAndProcessNextJob() {
			break
		}
	}
}

// getAndProcessNextJob job will process the next job of the queue job and returns if
// it needs to stop processing.
func (g *generic) getAndProcessNextJob() bool {
	// Get next job.
	nextJob, exit := g.queue.Get()
	if exit {
		return true
	}
	defer g.queue.Done(nextJob)
	key := nextJob.(string)

	// Our root span will start here.
	span := g.tracer.StartSpan("processJob")
	defer span.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	g.setRootSpanInfo(key, span)

	// Process the job. If errors then enqueue again.
	if err := g.processJob(ctx, key); err == nil {
		g.queue.Forget(key)
		g.setForgetSpanInfo(key, span, err)
	} else if g.queue.NumRequeues(key) < g.cfg.ProcessingJobRetries {
		// Job processing failed, requeue.
		g.logger.Warningf("error processing %s job (requeued): %v", key, err)
		g.queue.AddRateLimited(key)
		g.setReenqueueSpanInfo(key, span, err)
	} else {
		g.logger.Errorf("Error processing %s: %v", key, err)
		g.queue.Forget(key)
		g.setForgetSpanInfo(key, span, err)
	}

	return false
}

// processJob is where the real processing logic of the item is.
func (g *generic) processJob(ctx context.Context, key string) error {
	// Get the object
	obj, exists, err := g.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	// handle the object.
	if !exists { // Deleted resource from the cache.
		return g.handleDelete(ctx, key)
	}

	return g.handleAdd(ctx, key, obj.(runtime.Object))
}

func (g *generic) handleAdd(ctx context.Context, objKey string, obj runtime.Object) error {
	start := time.Now()

	// Create the span.
	pSpan := opentracing.SpanFromContext(ctx)
	span := g.tracer.StartSpan("handleAddObject", opentracing.ChildOf(pSpan.Context()))
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	// Set span data.
	ext.SpanKindConsumer.Set(span)
	span.SetTag(kubernetesObjectKeyKey, objKey)
	g.setCommonSpanInfo(span)
	span.LogKV(
		eventKey, "add",
		kubernetesObjectKeyKey, objKey,
	)

	// Handle the job.
	if err := g.handler.Add(ctx, obj); err != nil {
		ext.Error.Set(span, true) // Mark error as true.
		span.LogKV(
			eventKey, "error",
			messageKey, err,
		)

		g.metrics.IncResourceAddEventProcessedError(g.handlerName)
		g.metrics.ObserveDurationResourceAddEventProcessedError(g.handlerName, start)
		return err
	}
	g.metrics.IncResourceAddEventProcessedSuccess(g.handlerName)
	g.metrics.ObserveDurationResourceAddEventProcessedSuccess(g.handlerName, start)
	return nil
}

func (g *generic) handleDelete(ctx context.Context, objKey string) error {
	start := time.Now()

	// Create the span.
	pSpan := opentracing.SpanFromContext(ctx)
	span := g.tracer.StartSpan("handleDeleteObject", opentracing.ChildOf(pSpan.Context()))
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	// Set span data.
	ext.SpanKindConsumer.Set(span)
	span.SetTag(kubernetesObjectKeyKey, objKey)
	g.setCommonSpanInfo(span)
	span.LogKV(
		eventKey, "delete",
		kubernetesObjectKeyKey, objKey,
	)

	// Handle the job.
	if err := g.handler.Delete(ctx, objKey); err != nil {
		ext.Error.Set(span, true) // Mark error as true.
		span.LogKV(
			eventKey, "error",
			messageKey, err,
		)

		g.metrics.IncResourceDeleteEventProcessedError(g.handlerName)
		g.metrics.ObserveDurationResourceDeleteEventProcessedError(g.handlerName, start)
		return err
	}
	g.metrics.IncResourceDeleteEventProcessedSuccess(g.handlerName)
	g.metrics.ObserveDurationResourceDeleteEventProcessedSuccess(g.handlerName, start)
	return nil
}

func (g *generic) setCommonSpanInfo(span opentracing.Span) {
	ext.Component.Set(span, "kooper")
	span.SetTag(kooperControllerKey, g.cfg.Name)
	span.SetTag(controllerNameKey, g.cfg.Name)
	span.SetTag(controllerResyncKey, g.cfg.ResyncInterval)
	span.SetTag(controllerMaxRetriesKey, g.cfg.ProcessingJobRetries)
	span.SetTag(controllerConcurrentWorkersKey, g.cfg.ConcurrentWorkers)
}

func (g *generic) setRootSpanInfo(key string, span opentracing.Span) {
	numberRetries := g.queue.NumRequeues(key)

	// Try to set the namespace and resource name.
	if ns, name, err := cache.SplitMetaNamespaceKey(key); err == nil {
		span.SetTag(kubernetesObjectNSKey, ns)
		span.SetTag(kubernetesObjectNameKey, name)
	}

	g.setCommonSpanInfo(span)
	span.SetTag(kubernetesObjectKeyKey, key)
	span.SetTag(processedTimesKey, numberRetries+1)
	span.SetTag(processingRetryKey, numberRetries > 0)
	span.SetBaggageItem(kubernetesObjectKeyKey, key)
	ext.SpanKindConsumer.Set(span)
	span.LogKV(
		eventKey, "process_object",
		kubernetesObjectKeyKey, key,
	)
}

func (g *generic) setReenqueueSpanInfo(key string, span opentracing.Span, err error) {
	// Mark root span with error.
	ext.Error.Set(span, true)
	span.LogKV(
		eventKey, "error",
		messageKey, err,
	)

	rt := g.queue.NumRequeues(key)
	span.LogKV(
		eventKey, "reenqueued",
		retriesRemainingKey, g.cfg.ProcessingJobRetries-rt,
		retriesExecutedKey, rt,
		kubernetesObjectKeyKey, key,
	)
	span.LogKV(successKey, false)
}

func (g *generic) setForgetSpanInfo(key string, span opentracing.Span, err error) {
	success := true
	message := "object processed correctly"

	// Error data.
	if err != nil {
		// Mark root span with error.
		ext.Error.Set(span, true)
		span.LogKV(
			eventKey, "error",
			messageKey, err,
		)
		success = false
		message = "max number of retries reached after failing, forgetting object key"
	}

	span.LogKV(
		eventKey, "forget",
		messageKey, message,
		kubernetesObjectKeyKey, key,
	)
	span.LogKV(successKey, success)
}
