Controller tutorial
===================

In this tutorial we will learn how to create a controller using kooper. Yes, I know what you are thinking, kooper is more an operator library... but an operator as we described in the [concepts](concepts.md) is a controller on steroids and controllers are also fully supported in Kooper.

So... In this tutorial we will learn the pillars of the operator, the controller. The full controller is [here](https://github.com/spotahome/kooper/tree/master/examples/echo-pod-controller), we will go step by step but some of the code is *glue* or doesn't refer to kooper. 

Lets start!

## 01 - Description.

Our Controller will log all the add/delete events that occur to pods on a given namespace. Easy peasy... Lets call it  `echo-pod-controller` (yes, very original). The full controller is in [examples/echo-pod-controller](https://github.com/spotahome/kooper/tree/master/examples/echo-pod-controller).

### Structure.

The structure of the controller is very simple.

```bash
./examples/echo-pod-controller/
├── cmd
│   ├── flags.go
│   └── main.go
├── controller
│   ├── config.go
│   ├── echo.go
│   └── retrieve.go
├── log
│   └── log.go
└── service
    ├── service.go
    └── service_test.go
```

From this structure the important paths are `controller` where all the controller stuff will be, this is, creation, initialization... 

And `service` our domain logic.

The other ones are not so important for the tutorial, you should check the whole project to have in mind how is structured a full controller. `cmd` is the main program (flags, signal capturing, dependecy creation...). `log` is where our logger is.

### Unit testing.

Testing is important. As you see this project has very little unit tests. This is because of two things:

One, the project is very simple.

Two, you can trust Kubernetes and Kooper libraries, they are already tested, you don't need to test this, but you should test your domain logic (and if you want, main and glue code also). In this controller we just tested the service that has our domain logic (that is just a simple logging).

## 02 - Echo service.

First thigs first, we will implement our domain logic that doesn't know anything of our controller, in other words, our service will do the heavy stuff of the controller and what makes it special or different from other controllers.

Our controller is a logger just that.

```go
type Echo interface {
	EchoObj(prefix string, obj runtime.Object)
	EchoS(prefix string, s string)
}
```

We implemented that `Echo` service as `SimpleEcho` check it out on the [service file](https://github.com/spotahome/kooper/blob/master/examples/echo-pod-controller/service/service.go).

Now is time to use Kooper and leverage all the controller stuff.

## 03 - Controller configuration.

We need to implement our controller configuration. The controller is simple, it will need a namespace to know what pods should log and also a (re)synchronization period where the controller will receive all the pods again (apart from real time events).

[controller/config.go](https://github.com/spotahome/kooper/blob/master/examples/echo-pod-controller/controller/config.go)

```go
type Config struct {
	ResyncPeriod time.Duration
	Namespace    string
}
```

Simple.

Note we don't have validation, but you could set a method on the `Config` object to validate the configuration.

## 04 - Controller retriever.

Like we state on the basics, the retriever is the way the controller knows how to listen to resource events, this is, **list** (initial get and resyncs) and **watch** (real time events). And also know what is the kind of the object is listening to.

Our controller is for pods so our retriever will use the [Kubernetes client](https://github.com/kubernetes/client-go) to get the pods, for example:


```go
client.CoreV1().Pods(namespace).List(options)
```

Check the retriever [here](https://github.com/spotahome/kooper/blob/master/examples/echo-pod-controller/controller/retrieve.go)


## 05 - Controller handler.

As the name says, the handler is the place where kooper controller/operator will call when it has an event regarding the resource is listening with the retriever, in our case pods. 

Handler will receive events on:

* On the first iteration (when the controller starts) and on every resync (intervals) it will call as an `Add` so you get the full list of resources.
* On a resource deletion it will call `Delete`.
* On a resource update it will call `Add`

At first can look odd that an update on a resource calls `Add`. But we are getting a desired and eventual state of a resource, so doesn't matter if is new or old and has been updated, the reality is that our resource is in this state at this moment and we need to take actions or check previously before taking actions (imagine if we send an email, if we don't do a check we could end up with thousand of emails...).

In our case we don't bother to check if is new or old or if we have done something related with a previous event on the same resource. We just call our Echo Service.


[controller/echo.go](https://github.com/spotahome/kooper/blob/master/examples/echo-pod-controller/controller/echo.go)

```go
type handler struct {
	echoSrv service.Echo
}

func (h *handler) Add(_ context.Context, obj runtime.Object) error {
	h.echoSrv.EchoObj(addPrefix, obj)
	return nil
}
func (h *handler) Delete(_ context.Context, s string) error {
	h.echoSrv.EchoS(deletePrefix, s)
	return nil
}
```

## 06 - Controller.

We have all the pieces of the controller except the controller itself, but don't worry, Kooper gives you a controller implementation so you can glue all together and create a controller.

This can be found in [controller/echo.go](https://github.com/spotahome/kooper/blob/master/examples/echo-pod-controller/controller/echo.go) (is the same file of the handler).

We will go step by step:

```go
type Controller struct {
	controller.Controller

	config Config
	logger log.Logger
}
```

Controller is our controller, it has a logger, a controller configuration(step 03), and a kooper controller that will have all the required stuff to run a controller.

Our contructor starts by creating the dependencies to create `DefaultGeneric` kooper controller.

```go
ret := NewPodRetrieve(config.Namespace, k8sCli)
```

This is our retriever (step 04), the kubernetes client that we pass to the retriever constructor is created on the main and passed to the controller constructor (where we are and create this retriever).

```go
echoSrv := service.NewSimpleEcho(logger)
```
We create our service (step 01), this is for the handler.

```go
handler := &handler{echoSrv: echoSrv}
```

Then we create our simple handler that will have our service.

And... finally we create the controller!

```go
ctrl := controller.NewSequential(config.ResyncPeriod, handler, ret, nil, logger)
```

We are using a sequential controller constructor (`NewSequential`) from `"github.com/spotahome/kooper/operator/controller"` package. It receives a handler, a retriever, a logger and a resync period.

Wow, that was easy :)

## 07 - Finishing.

After all these steps we have a controller, now just depends how the main is organized or where you start the controller. You can check how is initialized the kubernetes client on the exmaple's [main]((https://github.com/spotahome/kooper/blob/master/examples/echo-pod-controller/cmd/main.go)), call our controller constructor and run it. But mainly is this:

```go
//...
ctrl, err := controller.New(cfg, k8sCli, logger)
//...
ctrl.Run(stopC)
```

The Run method receives a channel that when is closed all the controller stuff will be stopped.

