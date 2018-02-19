package controller

// Controller is the object that will implement the different kinds of controllers that will be running
// on the application.
type Controller interface {
	// Run runs the controller, it receives a channel that when receiving a signal it will stop the controller,
	// Run will block until it's stopped.
	Run(stopper <-chan struct{}) error
}
