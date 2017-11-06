package crd_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/client-go/pkg/api/v1"
	fcache "k8s.io/client-go/tools/cache/testing"

	"github.com/spotahome/redis-operator/mocks"
	"github.com/spotahome/redis-operator/pkg/crd"
)

type myCRD v1.Pod

func newTestWatcher(eh crd.EventHandler) (*crd.Watcher, *fcache.FakeControllerSource) {
	// Custom fake source of events
	stopC := make(chan int)
	source := fcache.NewFakeControllerSource()
	clientset := fake.NewSimpleClientset()

	return crd.NewWatcherWithSource(stopC, &myCRD{}, eh, source, clientset), source
}

func TestCRDWatcherAddEventReceived(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Mock a CRD add
	received := make(chan struct{})
	expectedCRD := &myCRD{}
	meh := &mocks.EventHandler{}
	meh.On("OnAdd", expectedCRD).Once().Run(func(_ mock.Arguments) {
		// On first add we are done, so notify that it has been processed
		received <- struct{}{}
	})

	w, fSrc := newTestWatcher(meh)

	// Get the fsource
	if assert.NotNil(w) {
		//Start watcher and close watcher on test finish
		w.Watch()
		defer func() {
			w.SignalChan <- 1
		}()

		// Send an add event
		fSrc.Add(expectedCRD)

		// Wait to process the events
	EventWaiter:
		for {
			select {
			case <-received:
				break EventWaiter
			case <-time.After(15 * time.Millisecond):
				require.Fail("Timeout receiving add event on event handler")
			}
		}
		// Check event handler calls
		meh.AssertExpectations(t)
	}
}

func TestCRDWatcherDeleteEventReceived(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Mock a CRD
	addRec := make(chan struct{})
	delRec := make(chan struct{})
	expectedCRD := &myCRD{}
	meh := &mocks.EventHandler{}
	meh.On("OnAdd", expectedCRD).Once().Run(func(_ mock.Arguments) {
		// On first add we are done, so notify that it has been processed
		addRec <- struct{}{}
	})
	meh.On("OnDelete", expectedCRD).Once().Run(func(_ mock.Arguments) {
		// On first add we are done, so notify that it has been processed
		delRec <- struct{}{}
	})

	w, fSrc := newTestWatcher(meh)

	// Get the fsource
	if assert.NotNil(w) {
		//Start watcher and close watcher on test finish
		w.Watch()
		defer func() {
			w.SignalChan <- 1
		}()

		// Send add an wait
		fSrc.Add(expectedCRD)

	AddWaiter:
		for {
			select {
			case <-addRec:
				break AddWaiter
			case <-time.After(15 * time.Millisecond):
				require.Fail("Timeout receiving add event on event handler")
			}
		}

		// Send the delete event
		fSrc.Delete(expectedCRD)
	DeleteWaiter:
		for {
			select {
			case <-delRec:
				break DeleteWaiter
			case <-time.After(15 * time.Millisecond):
				require.Fail("Timeout receiving delete event on event handler")
			}
		}
	}
}

func TestCRDWatcherUpdateEventReceived(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Mock a CRD
	addRec := make(chan struct{})
	upRec := make(chan struct{})
	expectedCRD := &myCRD{}
	meh := &mocks.EventHandler{}
	meh.On("OnAdd", expectedCRD).Once().Run(func(_ mock.Arguments) {
		// On first add we are done, so notify that it has been processed
		addRec <- struct{}{}
	})
	meh.On("OnUpdate", mock.AnythingOfType("*crd_test.myCRD"), expectedCRD).Once().Run(func(_ mock.Arguments) {
		// On first add we are done, so notify that it has been processed
		upRec <- struct{}{}
	})

	w, fSrc := newTestWatcher(meh)

	// Get the fsource
	if assert.NotNil(w) {
		//Start watcher and close watcher on test finish
		w.Watch()
		defer func() {
			w.SignalChan <- 1
		}()

		// Send add an wait
		fSrc.Add(expectedCRD)

	AddWaiter:
		for {
			select {
			case <-addRec:
				break AddWaiter
			case <-time.After(15 * time.Millisecond):
				require.Fail("Timeout receiving add event on event handler")
			}
		}

		// Send the delete event
		fSrc.Modify(expectedCRD)
	UpdateWaiter:
		for {
			select {
			case <-upRec:
				break UpdateWaiter
			case <-time.After(15 * time.Millisecond):
				require.Fail("Timeout receiving update event on event handler")
			}
		}
	}
}
