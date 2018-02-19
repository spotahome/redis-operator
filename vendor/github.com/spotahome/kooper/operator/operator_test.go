package operator_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/spotahome/kooper/log"
	mcontroller "github.com/spotahome/kooper/mocks/operator/controller"
	mresource "github.com/spotahome/kooper/mocks/operator/resource"
	"github.com/spotahome/kooper/operator"
	"github.com/spotahome/kooper/operator/controller"
	"github.com/spotahome/kooper/operator/resource"
)

func TestMultiOperatorInitialization(t *testing.T) {
	tests := []struct {
		name    string
		errInit bool
		expErr  bool
	}{
		{
			name:    "Calling multiple initializations should only ininitialize the once.",
			errInit: false,
			expErr:  false,
		},
		{
			name:    "Error on initialization should return error.",
			errInit: true,
			expErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			qCRD := 5

			var errInit error
			if test.errInit {
				errInit = errors.New("wanted error")
			}

			// Mocks.
			mcrds := []resource.CRD{}
			for i := 0; i < qCRD; i++ {
				crd := &mresource.CRD{}
				crd.On("Initialize").Once().Return(errInit)
				mcrds = append(mcrds, crd)
			}

			// Operator.
			op := operator.NewMultiOperator(mcrds, nil, log.Dummy)
			err := op.Initialize()

			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.NoError(op.Initialize(), "calling multiple times after initialization should not error")
			}
			for _, crd := range mcrds {
				mcrd := crd.(*mresource.CRD)
				mcrd.AssertExpectations(t)
			}
		})
	}
}

// controllerBehaviour is the way of controlling the beehaviour expected of a controller.
type controllerBehaviour struct {
	returnErr   bool
	returnAfter time.Duration
}

func createControllerMocks(cb []*controllerBehaviour) []controller.Controller {
	mctrls := make([]controller.Controller, len(cb))

	for i, b := range cb {
		// If we need to return an error return an error.
		var ctrlErr error
		if b.returnErr {
			ctrlErr = errors.New("wanted error")
		}

		ctrl := &mcontroller.Controller{}
		c := ctrl.On("Run", mock.Anything).Once().Return(ctrlErr)

		// If we need to wait anytime then delay the return.
		if b.returnAfter > 0 {
			c.After(b.returnAfter)
		}

		mctrls[i] = ctrl
	}
	return mctrls
}

func TestMultiOperatorRun(t *testing.T) {
	tests := []struct {
		name             string
		controllers      []*controllerBehaviour
		expErr           bool
		expEndOfOperator bool
	}{
		{
			name: "Running the operator should run without error if the controllers dont end.",
			controllers: []*controllerBehaviour{
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 9999 * time.Hour,
				},
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 9999 * time.Hour,
				},
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 9999 * time.Hour,
				},
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 9999 * time.Hour,
				},
			},
			expErr:           false,
			expEndOfOperator: false,
		},
		{
			name: "Running the operator should end without error if one of the controllers ends.",
			controllers: []*controllerBehaviour{
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 9999 * time.Hour,
				},
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 9999 * time.Hour,
				},
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 0, // The one that ends.
				},
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 9999 * time.Hour,
				},
			},
			expErr:           false,
			expEndOfOperator: true,
		},
		{
			name: "Running the operator should end with error if one of the controllers ends with an error.",
			controllers: []*controllerBehaviour{
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 9999 * time.Hour,
				},
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 9999 * time.Hour,
				},
				&controllerBehaviour{
					returnErr:   true,
					returnAfter: 0, // The one that ends.
				},
				&controllerBehaviour{
					returnErr:   false,
					returnAfter: 9999 * time.Hour,
				},
			},
			expErr:           true,
			expEndOfOperator: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			operatorStopperC := make(chan struct{})

			// Mocks.
			crd := &mresource.CRD{}
			crd.On("Initialize").Return(nil)
			mcrds := []resource.CRD{crd}

			mctrls := createControllerMocks(test.controllers)

			// Operator.
			op := operator.NewMultiOperator(mcrds, mctrls, log.Dummy)

			// Run in background and wait to the signals so it can test.
			errC := make(chan error)
			go func() {
				errC <- op.Run(operatorStopperC)
			}()

			// If the operator isn't expected to end we should end on our side.
			if !test.expEndOfOperator {
				// Wait a bit so the calls to controllers are done and then stop operator.
				time.Sleep(10 * time.Millisecond)
				close(operatorStopperC)
			}
			select {
			case err := <-errC:
				// Check.
				if test.expErr {
					assert.Error(err)
				} else {
					assert.NoError(err)
				}
			case <-time.After(1 * time.Second):
				assert.Fail("timeout waiting for controllers execution")
			}
		})
	}
}
