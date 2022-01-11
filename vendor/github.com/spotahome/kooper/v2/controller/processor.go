package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/spotahome/kooper/v2/log"
)

// processor knows how to process object keys.
type processor interface {
	Process(ctx context.Context, key string) error
}

// processorFunc a helper to create processors.
type processorFunc func(ctx context.Context, key string) error

func (p processorFunc) Process(ctx context.Context, key string) error { return p(ctx, key) }

// newIndexerProcessor returns a processor that processes a key that will get the kubernetes object
// from a cache called indexer were the kubernetes watch updates have been indexed and stored
// by the listerwatchers from the informers.
func newIndexerProcessor(indexer cache.Indexer, handler Handler) processor {
	return processorFunc(func(ctx context.Context, key string) error {
		// Get the object
		obj, exists, err := indexer.GetByKey(key)
		if err != nil {
			return err
		}

		if !exists {
			return nil
		}

		return handler.Handle(ctx, obj.(runtime.Object))
	})
}

var errRequeued = fmt.Errorf("requeued after receiving error")

// newRetryProcessor returns a processor that will delegate the processing of a key to the
// received processor, in case the processing/handling of this key fails it will add the key
// again to a queue if it has retrys pending.
//
// If the processing errored and has been retried, it will return a `errRequeued` error.
func newRetryProcessor(name string, queue blockingQueue, logger log.Logger, next processor) processor {
	return processorFunc(func(ctx context.Context, key string) error {
		err := next.Process(ctx, key)
		if err != nil {
			// Retry if possible.
			requeueErr := queue.Requeue(ctx, key)
			if requeueErr != nil {
				return fmt.Errorf("could not retry: %s: %w", requeueErr, err)
			}
			logger.WithKV(log.KV{"object-key": key}).Warningf("item requeued due to processing error: %s", err)
			return nil
		}

		return nil
	})
}

// newMetricsProcessor returns a processor that measures everything related with the processing logic.
func newMetricsProcessor(name string, mrec MetricsRecorder, next processor) processor {
	return processorFunc(func(ctx context.Context, key string) (err error) {
		defer func(t0 time.Time) {
			mrec.ObserveResourceProcessingDuration(ctx, name, err == nil, t0)
		}(time.Now())

		return next.Process(ctx, key)
	})
}
