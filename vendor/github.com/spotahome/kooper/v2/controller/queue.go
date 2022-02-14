package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"

	"github.com/spotahome/kooper/v2/log"
)

// blockingQueue is a queue that any of its implementations should
// implement a blocking get mechanism.
type blockingQueue interface {
	// Add will add an item to the queue.
	Add(ctx context.Context, item interface{})
	// Requeue will add an item to the queue in a requeue mode.
	// If doesn't accept requeueing or max requeue have been reached
	// it will return an error.
	Requeue(ctx context.Context, item interface{}) error
	// Get is a blocking operation, if the last object usage has not been finished (`done`)
	// being used it will block until this has been done.
	Get(ctx context.Context) (item interface{}, shutdown bool)
	// Done marks the item being used as done.
	Done(ctx context.Context, item interface{})
	// ShutDown stops the queue from accepting new jobs
	ShutDown(ctx context.Context)
	// Len returns the size of the queue.
	Len(ctx context.Context) int
}

var (
	errMaxRetriesReached = fmt.Errorf("max retries reached")
)

type rateLimitingBlockingQueue struct {
	maxRetries int
	queue      workqueue.RateLimitingInterface
}

func newRateLimitingBlockingQueue(maxRetries int, queue workqueue.RateLimitingInterface) blockingQueue {
	return rateLimitingBlockingQueue{
		maxRetries: maxRetries,
		queue:      queue,
	}
}

func (r rateLimitingBlockingQueue) Add(_ context.Context, item interface{}) {
	r.queue.Add(item)
}

func (r rateLimitingBlockingQueue) Requeue(_ context.Context, item interface{}) error {
	// If there was an error and we have retries pending then requeue.
	if r.queue.NumRequeues(item) < r.maxRetries {
		r.queue.AddRateLimited(item)
		return nil
	}

	r.queue.Forget(item)
	return errMaxRetriesReached
}

func (r rateLimitingBlockingQueue) Get(_ context.Context) (item interface{}, shutdown bool) {
	return r.queue.Get()
}

func (r rateLimitingBlockingQueue) Done(_ context.Context, item interface{}) {
	r.queue.Done(item)
}

func (r rateLimitingBlockingQueue) ShutDown(_ context.Context) {
	r.queue.ShutDown()
}

func (r rateLimitingBlockingQueue) Len(_ context.Context) int {
	return r.queue.Len()
}

// metricsQueue is a wrapper for a metrics measured queue.
type metricsBlockingQueue struct {
	mu            sync.Mutex
	name          string
	mrec          MetricsRecorder
	itemsQueuedAt map[interface{}]time.Time
	logger        log.Logger
	queue         blockingQueue
}

func newMetricsBlockingQueue(name string, mrec MetricsRecorder, queue blockingQueue, logger log.Logger) (blockingQueue, error) {
	// Register func/callback based metrics. These are controlled by the MetricsRecorder.
	err := mrec.RegisterResourceQueueLengthFunc(name, func(ctx context.Context) int { return queue.Len(ctx) })
	if err != nil {
		return nil, err
	}

	return &metricsBlockingQueue{
		name:          name,
		mrec:          mrec,
		itemsQueuedAt: map[interface{}]time.Time{},
		logger:        logger,
		queue:         queue,
	}, nil
}

func (m *metricsBlockingQueue) Add(ctx context.Context, item interface{}) {
	m.mu.Lock()
	if _, ok := m.itemsQueuedAt[item]; !ok {
		m.itemsQueuedAt[item] = time.Now()
	}
	m.mu.Unlock()

	m.mrec.IncResourceEventQueued(ctx, m.name, false)
	m.queue.Add(ctx, item)
}

func (m *metricsBlockingQueue) Requeue(ctx context.Context, item interface{}) error {
	m.mu.Lock()
	if _, ok := m.itemsQueuedAt[item]; !ok {
		m.itemsQueuedAt[item] = time.Now()
	}
	m.mu.Unlock()

	m.mrec.IncResourceEventQueued(ctx, m.name, true)
	return m.queue.Requeue(ctx, item)
}

func (m *metricsBlockingQueue) Get(ctx context.Context) (interface{}, bool) {
	// Here should get blocked, warning with the mutexes.
	item, shutdown := m.queue.Get(ctx)
	if shutdown {
		return item, shutdown
	}

	m.mu.Lock()
	queuedAt, ok := m.itemsQueuedAt[item]
	if ok {
		m.mrec.ObserveResourceInQueueDuration(ctx, m.name, queuedAt)
		delete(m.itemsQueuedAt, item)
	} else {
		m.logger.WithKV(log.KV{"object-key": item}).
			Infof("could not measure item because item is not present on metricsMeasuredQueue.itemsQueuedAt map")
	}
	m.mu.Unlock()

	return item, shutdown
}

func (m *metricsBlockingQueue) Done(ctx context.Context, item interface{}) {
	m.queue.Done(ctx, item)
}

func (m *metricsBlockingQueue) ShutDown(ctx context.Context) {
	m.queue.ShutDown(ctx)
}

func (m *metricsBlockingQueue) Len(ctx context.Context) int {
	// Measurement controlled by the metrics recorder, so is implemented in callback
	// mode, should be already registered, check factory. This is NOOP.
	return m.queue.Len(ctx)
}
