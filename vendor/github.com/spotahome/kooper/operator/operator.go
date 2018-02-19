package operator

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/operator/controller"
	"github.com/spotahome/kooper/operator/resource"
)

const (
	tiemoutInitialization = 1 * time.Minute
)

// Operator is a controller, at code level have almost same contract of behaviour
// but at a higher level it need to initialize some resources (usually CRDs) before
// start its execution.
type Operator interface {
	// Initialize knows how to initialize the resources.
	Initialize() error
	controller.Controller
}

// simpleOperator is an operator that initializes CRDs before starting
// the execution of controllers.
type simpleOperator struct {
	crds        []resource.CRD
	controllers []controller.Controller
	initialized bool
	running     bool
	stateMu     sync.Mutex
	logger      log.Logger
}

// NewOperator will return an operator that only manages one CRD
// and one Controller.
func NewOperator(crd resource.CRD, ctrlr controller.Controller, logger log.Logger) Operator {
	return NewMultiOperator([]resource.CRD{crd}, []controller.Controller{ctrlr}, logger)
}

// NewMultiOperator returns an operator that has multiple CRDs and controllers.
func NewMultiOperator(crds []resource.CRD, ctrlrs []controller.Controller, logger log.Logger) Operator {
	return &simpleOperator{
		crds:        crds,
		controllers: ctrlrs,
		logger:      logger,
	}
}

// Initialize will initializer all the CRDs and return. Satisfies Operator interface.
func (s *simpleOperator) Initialize() error {
	if s.isInitialized() {
		return nil
	}

	// Initialize CRDs.
	var g errgroup.Group
	for _, crd := range s.crds {
		crd := crd
		g.Go(func() error {
			return crd.Initialize()
		})
	}

	// Wait until everything is initialized.
	errC := make(chan error)
	go func() {
		errC <- g.Wait()
	}()

	select {
	case err := <-errC:
		if err != nil {
			return err
		}
	case <-time.After(tiemoutInitialization):
		return fmt.Errorf("timeout initializing operator")
	}

	// All ok, we are ready to run.
	s.logger.Infof("operator initialized")
	s.setInitialized(true)
	return nil
}

// Run will run the operator (a.k.a) the controllers and Initialize the CRDs.
// It's a blocking operation. Satisfies Operator interface. The client that uses an operator
// has the responsibility of closing the stop channel if the operator ends execution
// unexpectly so all the goroutines (controllers running) end its execution
func (s *simpleOperator) Run(stopC <-chan struct{}) error {
	if s.isRunning() {
		return fmt.Errorf("operator already running")
	}

	s.logger.Infof("starting operator")
	s.setRunning(true)
	defer s.setRunning(false)

	if err := s.Initialize(); err != nil {
		return err
	}

	errC := make(chan error)
	go func() {
		errC <- s.runAllControllers(stopC)
	}()

	// When stop signal is received we must stop or when an error is received from
	// one of the controllers.
	select {
	case err := <-errC:
		if err != nil {
			return err
		}
	case <-stopC:
	}
	s.logger.Infof("stopping operator")
	return nil
}

// runAllControllers will run controllers and block execution.
func (s *simpleOperator) runAllControllers(stopC <-chan struct{}) error {
	errC := make(chan error)

	for _, ctrl := range s.controllers {
		ctrl := ctrl
		go func() {
			errC <- ctrl.Run(stopC)
		}()
	}

	// Wait until any of the  controllers end execution. All the controllers should be executing so
	// if we receive any result (error or not) we should return an error
	select {
	case <-stopC:
		return nil
	case err := <-errC:
		if err != nil {
			return fmt.Errorf("a controller ended with an error: %s", err)
		}
		s.logger.Warningf("a controller stopped execution without error before operator received stop signal, stopping operator.")
		return nil
	}
}

func (s *simpleOperator) isInitialized() bool {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	return s.initialized
}

func (s *simpleOperator) setInitialized(value bool) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	s.initialized = value
}

func (s *simpleOperator) isRunning() bool {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	return s.running
}

func (s *simpleOperator) setRunning(value bool) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	s.running = value
}
