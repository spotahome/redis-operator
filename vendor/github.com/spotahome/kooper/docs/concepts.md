
Start
=====

Kooper motivation: Help creating Kubernetes operators and controllers by giving a framework, libraries or a set of tools (pick the name that you prefer).

If you are new to kooper first get familiar with the basic concepts and design of this toolkit, after that you will be redirected to some simple operators and controllers explained so you can see what can you do with this toolkit.
	
## Concepts

These will be the pieces of the framework required to implement and know to create an operator.
The required pieces you need to implement and know to create an operator are:
* `controller`: Take actions to comply the needs of resources.
* `retriever`: Knows how to list, watch and create a void object.
* `CRD`: A retriever that knows how to Initialize.
* `Handler`: Will be called when were events in the resources.

### Controller

A Kubernetes controller is something that listens/watch the status of a Kubernetes resource and takes actions so it meets the state required by the resource. Example:

I listen to a `replicaset` resource that wants N instances of X `pod`. My controller should take whatever action it considers so the cluster has N instances of X. In this case It will create N `pod` resources.

In this framework it is defined as an interface [here](https://github.com/spotahome/kooper/tree/master/operator/controller)

```go
package controller

type Controller interface {
	Run(stopper <-chan struct{}) error
}
```

### Retriever

In kubernetes world the way of retrieving resources is listing or watching (streaming events) them. In Kubernetes client libraries they are called [listerwatchers](https://github.com/kubernetes/client-go/blob/1b825e3a786379cb2ae2edc98a39e9c8cd68ee3c/tools/cache/listwatch.go#L35-L41). They know how to list and watch a resource kind.

In this framework we have the concept of `Retriever` It knows how to list, watch and create a void object
so we can know the kind of the Retriever by inspecting the object. It is defined [here](https://github.com/spotahome/kooper/tree/master/operator/retrieve)

```go
type Retriever interface {
	GetListerWatcher() cache.ListerWatcher
	GetObject() runtime.Object
}
```

### CRD

A `crd` is a Custom Resource Definition (the evolution of the TPRs). It's a Kubernetes resource that is not a Kubernetes base resource (`pod`, `deployment`, `secret`...). It's used so we (as Kubernetes users) can create our custom resources (in the end a manifest/spec/definition) and use new resource kinds inside the cluster.

In this framework we have the concept of `CRD`. It's just a retriever that knows how to Initialize.
The Initialize exists because when you want to use it in a Kubernetes cluster you need to ensure that the CRD is previously present (registered).

```go
type CRD interface {
	retrieve.Retriever
	Initialize() error
}
```

### Handler

The Handler is where our operator/controller logic will be placed. In other words, every time the resource we are *listwatching* for its events (delete, create, changed) the handler will be called.

```go
type Handler interface {
	Add(context.Context, runtime.Object) error
	Delete(context.Context, string) error
}
```

Maybe you have some doubts like... 

* Where is the update method?
In the Kuberntes world there is the concept of state and eventual consistency and reconciliation loops. In a few words, You don't need to know
if the resource is new or not, only that the state should be this and eventually that state will be real in the cluster.

* What happens if it errors my handling?
The event will be requeued for a new handling in the future until it rate limits the maximum times allowed (if this isn't rate limited you could get stuck forever handling same resources)

* A context as parameter?
The context can be ignored if you don't need it at all, but if tracing is active the context will have the parent span.


### Operator

All the concepts described above are glued together and create an operator. An [operator](https://coreos.com/operators/) its a concept invented by the CoreOS devs that is basically a regular controller that automates operator (a person) actions, so we manage CRDs to represent resources that our controllers will manage them. Example:

I need to deploy a Prometheus, set up configuration... Instead of doing this myself (as a human operator) I create a new resource kind called `Prometheus`, or in other words a CRD called `Prometheus` with different options and create a controller that knows how to set the required state for our Prometheus CRD (this means deploying, configuring, backups, updating... a Promethus instance.).

This way we just automated repetitive stuff, remove toil, errors...

In this framework the concept operator is: controller(s) + CRD(s) = Operator. In other words an Operator initializes one or more CRDs and the runs one or more controllers related to one or more CRDs previously initialized.

```go
type Operator interface {
	Initialize() error
	controller.Controller
}
```

## Usage

In the end when using this framework or toolkit you need to know the concepts, but you will not use
all of them directly. For example we provide ways of bootstraping controllers and operators with already implemented utils.

## Next

At this moment you have the basic knowledge or the pillars that sustain this framework/toolkit, the next
thing you need to know is how to create and run stuff built with this toolkit:

* [Controller tutorial.](controller-tutorial.md)
* [Operator tutorial.](operator-tutorial.md)
