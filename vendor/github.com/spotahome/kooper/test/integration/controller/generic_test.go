// +build integration

package controller_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/monitoring/metrics"
	"github.com/spotahome/kooper/operator/controller"
	"github.com/spotahome/kooper/operator/handler"
	"github.com/spotahome/kooper/operator/retrieve"
)

const (
	noResync             = 0
	controllerRunTimeout = 10 * time.Second
	// this delta is the max duration delta used on the assertion of controller handling, this is required because
	// the controller requires some millisecond to bootstrap and sync.
	maxAssertDurationDelta = 500 * time.Millisecond
)

func returnPodList(q int) *corev1.PodList {
	items := make([]corev1.Pod, q)

	for i := 0; i < q; i++ {
		items[i] = corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("pod%d", i),
			},
		}
	}
	return &corev1.PodList{
		Items: items,
	}
}

// runTimedController will run a controller that will handle multiple events and will return the duration
// how long it took to process all the events. each handled event will take the desired ammount of time.
func runTimedController(sleepDuration time.Duration, concurrencyLevel int, numberOfEvents int, t *testing.T) time.Duration {
	assert := assert.New(t)

	// Create the faked retriever that will only return N pods.
	podList := returnPodList(numberOfEvents)
	r := &retrieve.Resource{
		Object: &corev1.Pod{},
		ListerWatcher: &cache.ListWatch{
			ListFunc: func(_ metav1.ListOptions) (runtime.Object, error) {
				return podList, nil
			},
			WatchFunc: func(_ metav1.ListOptions) (watch.Interface, error) {
				return watch.NewFake(), nil
			},
		},
	}

	// Create the handler that will wait on each event T duration and will
	// end when all the wanted quantity of events have been processed.
	var wg sync.WaitGroup
	wg.Add(numberOfEvents)

	h := &handler.HandlerFunc{
		AddFunc: func(_ context.Context, _ runtime.Object) error {
			time.Sleep(sleepDuration)
			wg.Done()
			return nil
		},
		DeleteFunc: func(_ context.Context, _ string) error {
			assert.Fail("delete events should not be used on this test")
			return nil
		},
	}

	// Create the controller type depending on the concurrency level.
	var ctrl controller.Controller
	var err error
	if concurrencyLevel < 2 {
		ctrl = controller.NewSequential(noResync, h, r, metrics.Dummy, log.Dummy)
	} else {
		ctrl, err = controller.NewConcurrent(concurrencyLevel, noResync, h, r, metrics.Dummy, log.Dummy)
		if !assert.NoError(err) {
			return 0
		}
	}

	// Run handling
	stopC := make(chan struct{})
	defer close(stopC)
	start := time.Now()
	go func() {
		assert.NoError(ctrl.Run(stopC))
	}()

	// Wait until the finish event is received (wait until all events processed), it has a big timeout (it's an integration test).
	finishC := make(chan struct{})
	go func() {
		wg.Wait()
		close(finishC)
	}()
	select {
	case <-time.After(controllerRunTimeout):
		assert.Fail("timeout waiting for controller finish processing events")
	case <-finishC:
	}

	// Return result duration of all the handling.
	return time.Now().Sub(start)
}

func TestGenericControllerSequentialVSConcurrent(t *testing.T) {
	tests := []struct {
		name                  string
		numberOfEvents        int
		concurrencyLevel      int
		sleepDuration         time.Duration
		expSequentialDuration time.Duration
		expConcurrentDuration time.Duration
	}{
		{
			name:                  "100 events with 10ms handling latency should take 1s sequentially and 200ms with 5 workers",
			numberOfEvents:        100,
			concurrencyLevel:      5,
			sleepDuration:         10 * time.Millisecond,
			expSequentialDuration: 1 * time.Second,
			expConcurrentDuration: 200 * time.Millisecond,
		},
		{
			name:                  "100 events with 20ms handling latency should take 2s sequentially and 400ms with 5 workers",
			numberOfEvents:        100,
			concurrencyLevel:      5,
			sleepDuration:         20 * time.Millisecond,
			expSequentialDuration: 2 * time.Second,
			expConcurrentDuration: 400 * time.Millisecond,
		},
		{
			name:                  "100 events with 30ms handling latency should take 3s sequentially and 600ms with 5 workers",
			numberOfEvents:        100,
			concurrencyLevel:      5,
			sleepDuration:         30 * time.Millisecond,
			expSequentialDuration: 3 * time.Second,
			expConcurrentDuration: 600 * time.Millisecond,
		},
		{
			name:                  "100 events with 40ms handling latency should take 4s sequentially and 800ms with 5 workers",
			numberOfEvents:        100,
			concurrencyLevel:      5,
			sleepDuration:         40 * time.Millisecond,
			expSequentialDuration: 4 * time.Second,
			expConcurrentDuration: 800 * time.Millisecond,
		},
		{
			name:                  "100 events with 50ms handling latency should take 5s sequentially and 1s with 5 workers",
			numberOfEvents:        100,
			concurrencyLevel:      5,
			sleepDuration:         50 * time.Millisecond,
			expSequentialDuration: 5 * time.Second,
			expConcurrentDuration: 1 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			gotSecDuration := runTimedController(test.sleepDuration, 1, test.numberOfEvents, t)
			gotConcDuration := runTimedController(test.sleepDuration, test.concurrencyLevel, test.numberOfEvents, t)

			// Check if the expected time is correct. check expected and got duration  are within a max delta (usually controller bootstrapping/sync).
			assert.InDelta(test.expSequentialDuration, gotSecDuration, float64(maxAssertDurationDelta))
			assert.InDelta(test.expConcurrentDuration, gotConcDuration, float64(maxAssertDurationDelta))
		})
	}
}
