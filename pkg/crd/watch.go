package crd

import (
	"time"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/spotahome/redis-operator/pkg/clock"
)

const onStatusLoopInterval = 3 * time.Minute

// EventHandler interface has the logic about handling the different events of a k8s resource
type EventHandler interface {
	OnAdd(obj interface{})
	OnUpdate(oldObj, newObj interface{})
	OnDelete(obj interface{})
	OnStatus()
}

// Watch is an interface that watcher for a k8s resource
type Watch interface {
	Watch()
}

// Watcher implements Watcher interface using k8s CRDs
type Watcher struct {
	SignalChan   chan int
	EventHandler EventHandler

	// TODO: StopWatch method
	stop chan struct{}
	// source is the listerwatcher where the informer will listen for events;
	// I'ts been decoupled in order so we can mock the vents
	source cache.ListerWatcher
	// controller is an informer where the events will be received an it will
	// execute the required logic from the event handlers
	controller cache.Controller
}

// NewWatcher returns a new CRD watcher
func NewWatcher(clientset cache.Getter, signalChan chan int, apiName string, object runtime.Object, eventHandler EventHandler, k8sclient apiextensionsclient.Interface) *Watcher {
	// Create the informer where wi will receive the events
	source := cache.NewListWatchFromClient(clientset, apiName, v1.NamespaceAll, fields.Everything())
	return NewWatcherWithSource(signalChan, object, eventHandler, source, k8sclient)
}

// NewWatcherWithSource returns a new CRD watcher with a custom source of events
func NewWatcherWithSource(signalChan chan int, object runtime.Object, eventHandler EventHandler, source cache.ListerWatcher, clientset apiextensionsclient.Interface) *Watcher {
	t := &Watcher{
		SignalChan:   signalChan,
		EventHandler: eventHandler,
		source:       source,

		stop: make(chan struct{}),
	}

	// Create the controller where the events received from the events will be processed

	t.controller = NewInformer(t.source, object, 0, t.EventHandler, clock.Base(), clientset)

	return t
}

// Watch will start the event handling loop
func (t *Watcher) Watch() {
	go t.controller.Run(t.stop)

	// Stop watcher when received channel receives the signal or closed
	go func() {
		select {
		case <-t.SignalChan:
			close(t.stop)
		}
	}()
}

// Informer is wrapper of controller to be able to get OnStatus calls
type Informer struct {
	cache.Controller
	EventHandler EventHandler
	clock        clock.Clock
	clientset    apiextensionsclient.Interface
}

// NewInformer is the constructor of the Informer struct
func NewInformer(lw cache.ListerWatcher, objType runtime.Object, resyncPeriod time.Duration, eh EventHandler, clock clock.Clock, clientset apiextensionsclient.Interface) *Informer {
	_, controller := cache.NewInformer(
		lw,
		objType,
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    eh.OnAdd,
			UpdateFunc: eh.OnUpdate,
			DeleteFunc: eh.OnDelete,
		})

	return &Informer{
		Controller:   controller,
		EventHandler: eh,
		clock:        clock,
		clientset:    clientset,
	}
}

// Run starts the CRD, with the watchers and the checker loop
func (t *Informer) Run(stopCh <-chan struct{}) {
	go func() {
		ticker := t.clock.NewTicker(onStatusLoopInterval)
		for range ticker.C {
			t.EventHandler.OnStatus()
		}
	}()
	t.Controller.Run(stopCh)
}
