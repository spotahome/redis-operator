package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/spotahome/kooper/v2/controller/leaderelection"
	"github.com/spotahome/kooper/v2/log"
)

var (
	// ErrControllerNotValid will be used when the controller has invalid configuration.
	ErrControllerNotValid = errors.New("controller not valid")
)

// Controller is the object that will implement the different kinds of controllers that will be running
// on the application.
type Controller interface {
	// Run runs the controller and blocks until the context is `Done`.
	Run(ctx context.Context) error
}

// Config is the controller configuration.
type Config struct {
	// Handler is the controller handler.
	Handler Handler
	// Retriever is the controller retriever.
	Retriever Retriever
	// Leader elector will be used to use only one instance, if no set it will be
	// leader election will be ignored
	LeaderElector leaderelection.Runner
	// MetricsRecorder will record the controller metrics.
	MetricsRecorder MetricsRecorder
	// Logger will log messages of the controller.
	Logger log.Logger

	// name of the controller.
	Name string
	// ConcurrentWorkers is the number of concurrent workers the controller will have running processing events.
	ConcurrentWorkers int
	// ResyncInterval is the interval the controller will process all the selected resources.
	ResyncInterval time.Duration
	// ProcessingJobRetries is the number of times the job will try to reprocess the event before returning a real error.
	ProcessingJobRetries int
	// DisableResync will disable resyncing, if disabled the controller only will react on event updates and resync
	// all when it runs for the first time.
	// This is useful for secondary resource controllers (e.g pod controller of a primary controller based on deployments).
	DisableResync bool
}

func (c *Config) setDefaults() error {
	if c.Name == "" {
		return fmt.Errorf("a controller name is required")
	}

	if c.Handler == nil {
		return fmt.Errorf("a handler is required")
	}

	if c.Retriever == nil {
		return fmt.Errorf("a retriever is required")
	}

	if c.Logger == nil {
		c.Logger = log.NewStd(false)
		c.Logger.Warningf("no logger specified, fallback to default logger, to disable logging use a explicit Noop logger")
	}
	c.Logger = c.Logger.WithKV(log.KV{
		"service":       "kooper.controller",
		"controller-id": c.Name,
	})

	if c.MetricsRecorder == nil {
		c.MetricsRecorder = DummyMetricsRecorder
		c.Logger.Warningf("no metrics recorder specified, disabling metrics")
	}

	if c.ConcurrentWorkers <= 0 {
		c.ConcurrentWorkers = 3
	}

	if c.ResyncInterval <= 0 {
		c.ResyncInterval = 3 * time.Minute
	}

	if c.DisableResync {
		c.ResyncInterval = 0 // 0 == resync disabled.
	}

	if c.ProcessingJobRetries < 0 {
		c.ProcessingJobRetries = 0
	}

	return nil
}

// generic controller is a controller that can be used to create different kind of controllers.
type generic struct {
	queue     blockingQueue             // queue will have the jobs that the controller will get and send to handlers.
	informer  cache.SharedIndexInformer // informer will notify be inform us about resource changes.
	processor processor                 // processor will call the user handler (logic).

	running   bool
	runningMu sync.Mutex
	cfg       Config
	metrics   MetricsRecorder
	leRunner  leaderelection.Runner
	logger    log.Logger
}

func listerWatcherFromRetriever(ret Retriever) cache.ListerWatcher {
	// TODO(slok): pass context when Kubernetes updates its ListerWatchers ¯\_(ツ)_/¯.
	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return ret.List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return ret.Watch(context.TODO(), options)
		},
	}
}

// New creates a new controller that can be configured using the cfg parameter.
func New(cfg *Config) (Controller, error) {
	// Sets the required default configuration.
	err := cfg.setDefaults()
	if err != nil {
		return nil, fmt.Errorf("could no create controller: %w: %v", ErrControllerNotValid, err)
	}

	// Create the queue that will have our received job changes.
	queue := newRateLimitingBlockingQueue(
		cfg.ProcessingJobRetries,
		workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	)

	// Measure the queue.
	queue, err = newMetricsBlockingQueue(
		cfg.Name,
		cfg.MetricsRecorder,
		queue,
		cfg.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("could not measure the queue: %w", err)
	}

	// store is the internal cache where objects will be store.
	store := cache.Indexers{}
	lw := listerWatcherFromRetriever(cfg.Retriever)
	informer := cache.NewSharedIndexInformer(lw, nil, cfg.ResyncInterval, store)

	// Set up our informer event handler.
	// Objects are already in our local store. Add only keys/jobs on the queue so they can re processed
	// afterwards.
	informer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				cfg.Logger.Warningf("could not add item from 'add' event to queue: %s", err)
				return
			}
			queue.Add(context.TODO(), key)
		},
		UpdateFunc: func(_ interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				cfg.Logger.Warningf("could not add item from 'update' event to queue: %s", err)
				return
			}
			queue.Add(context.TODO(), key)
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				cfg.Logger.Warningf("could not add item from 'delete' event to queue: %s", err)
				return
			}
			queue.Add(context.TODO(), key)
		},
	}, cfg.ResyncInterval)

	// Create processing chain: processor(+middlewares) -> handler(+middlewares).
	processor := newIndexerProcessor(informer.GetIndexer(), cfg.Handler)
	if cfg.ProcessingJobRetries > 0 {
		processor = newRetryProcessor(cfg.Name, queue, cfg.Logger, processor)
	}
	processor = newMetricsProcessor(cfg.Name, cfg.MetricsRecorder, processor)

	// Create our generic controller object.
	return &generic{
		queue:     queue,
		informer:  informer,
		metrics:   cfg.MetricsRecorder,
		processor: processor,
		leRunner:  cfg.LeaderElector,
		cfg:       *cfg,
		logger:    cfg.Logger,
	}, nil
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
func (g *generic) Run(ctx context.Context) error {
	// Check if leader election is required.
	if g.leRunner != nil {
		return g.leRunner.Run(func() error {
			return g.run(ctx)
		})
	}

	return g.run(ctx)
}

// run is the real run of the controller.
func (g *generic) run(ctx context.Context) error {
	if g.isRunning() {
		return fmt.Errorf("controller already running")
	}

	g.logger.Infof("starting controller")
	// Set state of controller.
	g.setRunning(true)
	defer g.setRunning(false)

	// Shutdown when Run is stopped so we can process the last items and the queue doesn't
	// accept more jobs.
	defer g.queue.ShutDown(ctx)

	// Run the informer so it starts listening to resource events.
	go g.informer.Run(ctx.Done())

	// Wait until our store, jobs... stuff is synced (first list on resource, resources on store and jobs on queue).
	if !cache.WaitForCacheSync(ctx.Done(), g.informer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	// Start our resource processing worker, if finishes then restart the worker. The workers should
	// not end.
	for i := 0; i < g.cfg.ConcurrentWorkers; i++ {
		go func() {
			wait.Until(g.runWorker, time.Second, ctx.Done())
		}()
	}

	// Block while running our workers in a continuous way (and re run if they fail). But
	// when stop signal is received we must stop.
	<-ctx.Done()
	g.logger.Infof("stopping controller")

	return nil
}

// runWorker will start a processing loop on event queue.
func (g *generic) runWorker() {
	for {
		// Process next queue job, if needs to stop processing it will return true.
		if g.processNextJob() {
			break
		}
	}
}

// processNextJob job will process the next job of the queue job and returns if
// it needs to stop processing.
//
// If the queue has been closed then it will end the processing.
func (g *generic) processNextJob() bool {
	ctx := context.Background()

	// Get next job.
	nextJob, exit := g.queue.Get(ctx)
	if exit {
		return true
	}

	defer g.queue.Done(ctx, nextJob)
	key := nextJob.(string)

	// Process the job.
	err := g.processor.Process(ctx, key)

	logger := g.logger.WithKV(log.KV{"object-key": key})
	switch {
	case err == nil:
		logger.Debugf("object processed")
	case errors.Is(err, errRequeued):
		logger.Warningf("error on object processing, retrying: %v", err)
	default:
		logger.Errorf("error on object processing: %v", err)
	}

	return false
}
