# Kooper [![Build Status][travis-image]][travis-url] [![Go Report Card][goreport-image]][goreport-url] [![GoDoc][godoc-image]][godoc-url]

Kooper is a simple Go library to create Kubernetes [operators](https://coreos.com/operators/) and [controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md).

## What is Kooper?

Kooper is a set of utilities packed as a library or framework to easily create Kubernetes controllers and operators.

There is a little of discussion of what a controller and what an operator is, see [here](https://stackoverflow.com/questions/47848258/kubernetes-controller-vs-kubernetes-operator) for more information.

In Kooper the concepts of controller an operator are very simple, a controller controls the state of a resource in Kubernetes, and an operator is a controller that initializes custom resources (CRD) and controls the state of this custom resource.

## Features

- Easy and decoupled library.
- Well structured and a clear API.
- Remove all duplicated code from every controller and operator.
- Uses the tooling already created by Kubernetes.
- Remove complexity from operators and controllers so the focus is on domain logic.
- Easy to mock and extend functionality (Go interfaces!).
- Only support CRD, no TPR support (Kubernetes >=1.7).
- Controller metrics (with a [Grafana dashboard][grafana-dashboard] based on Prometheus metrics backend).
- Leader election.
- Tracing with [Opentracing][opentracing-url].

## Example

It can be seen how easy is to develop a controller or an operator in kooper looking at the [documentation](docs).

This is a simple pod log controller example ([full running example here](https://github.com/spotahome/kooper/blob/master/examples/onefile-echo-pod-controller/main.go)):

```go
// Initialize resources like logger and kubernetes client
//...

// Create our retriever so the controller knows how to get/listen for pod events.
retr := &retrieve.Resource{
    Object: &corev1.Pod{},
    ListerWatcher: &cache.ListWatch{
        ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
            return k8scli.CoreV1().Pods("").List(options)
        },
        WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
            return k8scli.CoreV1().Pods("").Watch(options)
        },
    },
}

// Our domain logic that will print every add/sync/update and delete event.
hand := &handler.HandlerFunc{
    AddFunc: func(_ context.Context, obj runtime.Object) error {
        pod := obj.(*corev1.Pod)
        log.Infof("Pod added: %s/%s", pod.Namespace, pod.Name)
        return nil
    },
    DeleteFunc: func(_ context.Context, s string) error {
        log.Infof("Pod deleted: %s", s)
        return nil
    },
}

// Create the controller that will refresh every 30 seconds.
ctrl := controller.NewSequential(30*time.Second, hand, retr, nil, log)
stopC := make(chan struct{})
if err := ctrl.Run(stopC); err != nil {
    log.Errorf("error running controller: %s", err)
    os.Exit(1)
}
os.Exit(0)
```

The above shows that is very easy to get a controller working in less than 100 lines of code. How it works can be demonstrated by running the controller from this repository.

```bash
go run ./examples/onefile-echo-pod-controller/main.go
```

or directly using docker (Note you are binding kubernetes configuration on the container so kooper can connect to kubernetes cluster).

```bash
docker run \
    --rm -it \
    -v ${HOME}/.kube:/root/.kube:ro \
    golang:1.10 \
    /bin/bash -c "go get github.com/spotahome/kooper/... && cd /go/src/github.com/spotahome/kooper && go run ./examples/onefile-echo-pod-controller/main.go"
```

## Motivation

The state of art in the operators/controllers moves fast, a lot of new operators are being published every day. Most of them have the same "infrastructure" code refering Kubernetes operators/controllers and bootstrapping a new operator can be slow or repetitive.

At this moment there is no standard, although there are some projects like [rook operator kit](https://github.com/rook/operator-kit) or [Giantswarm operator kit](https://github.com/giantswarm/operatorkit) that are trying to create it.

Spotahome studied these projects before developing Kooper and they didn't fit the requirements:

- Clear and maintanable code.
- Easy to test and mock.
- Well tested library.
- Easy and clear programming API.
- Good abstraction and structure to focus on domain logic (the meat of the controller).
- Reduce complexity in all the operators and controllers that use the library.
- Not only operators, controllers as first class citizen also.

## Installing

Any dependency manager can get Kooper or directly with go get the latest version:

```bash
go get -u github.com/spotahome/kooper
```

## Using Kooper as a dependency

Managing a project that uses different kubernetes libs as dependencies can be tricky at first because of the different versions of all these libraries(apimachinery, client-go, api...). [Here][dependency-example] you have an example of how would you use kooper as a dependency in a project and setting the kubernetes libraries to the version that you want along with kooper (using [dep][dep-project]).

## Compatibility matrix

|             | Kubernetes 1.8 | Kubernetes 1.9 | Kubernetes 1.10 | Kubernetes 1.11 | Kubernetes 1.12 |
| ----------- | -------------- | -------------- | --------------- | --------------- | --------------- |
| kooper 0.1  | ✓              | ✓              | ?               | ?               | ?               |
| kooper 0.2  | ✓              | ✓              | ?               | ?               | ?               |
| kooper 0.3  | ?              | ?+             | ✓               | ?               | ?               |
| kooper 0.4  | ?              | ?+             | ✓               | ?               | ?               |
| kooper 0.5  | ?              | ?              | ?+              | ✓               | ?               |
| kooper HEAD | ?              | ?              | ?+              | ?+              | ✓               |

Based on this matrix Kooper needs different versions of Kubernetes dependencies.

An example would be. If the cluster that will use kooper operators/controllers is a 1.9.x Kubernetes cluster, the version of kooper would be `0.2.x` and the kubernetes libraries in `1.9.x` compatibility style. For example the project that uses kooper as a dependency, its dep file would be something like this:

```yaml
...

[[override]]
  name = "k8s.io/api"
  version = "kubernetes-1.9.6"

[[override]]
  name = "k8s.io/apimachinery"
  version = "kubernetes-1.9.6"

[[override]]
  name = "k8s.io/client-go"
  version = "kubernetes-1.9.6"

[[constraint]]
  name = "github.com/spotahome/kooper"
  version = "0.2.0"

...
```

## Documentation

Kooper comes with different topics as documentation.

### Core

- [Basic concepts](docs/concepts.md)
- [Controller tutorial](docs/controller-tutorial.md)
- [Operator tutorial](docs/operator-tutorial.md)

### Other

- [Metrics](docs/metrics.md)
- [Logger](docs/logger.md)
- [Leader election](docs/leader-election.md)
- [Tracing](docs/tracing.md)

The starting point would be to check the [concepts](docs/concepts.md) and then continue with the controller and operator tutorials.

## Who is using kooper

- [redis-operator](https://github.com/spotahome/redis-operator)
- [node-labeler-operator](https://github.com/barpilot/node-labeler-operator)
- [source-ranges-controller](https://github.com/jeffersongirao/source-ranges-controller)

[travis-image]: https://travis-ci.org/spotahome/kooper.svg?branch=master
[travis-url]: https://travis-ci.org/spotahome/kooper
[goreport-image]: https://goreportcard.com/badge/github.com/spotahome/kooper
[goreport-url]: https://goreportcard.com/report/github.com/spotahome/kooper
[godoc-image]: https://godoc.org/github.com/spotahome/kooper?status.svg
[godoc-url]: https://godoc.org/github.com/spotahome/kooper
[dependency-example]: https://github.com/slok/kooper-as-dependency
[dep-project]: https://github.com/golang/dep
[opentracing-url]: http://opentracing.io/
[grafana-dashboard]: https://grafana.com/dashboards/7082
