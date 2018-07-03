# Metrics

Kooper comes with metrics support, this means that when you use kooper to bootstrap your operator or controller you have the possibility of instrumenting your controller for free.

## Backends

At this moment these are the supported backends:

* Prometheus.

## Custom backend

ALthough Kooper supports by default some of the de-facto standard instrumenting backends, you could create your own backend, you just need to implement the provided [interface][metrics-interface] that is named as `Recorder`:

```golang
type Recorder interface {
    ...
}
```

## Measured metrics

The measured metrics are:

* Number of delete and add queued events (to be processed).
* Number of delete and add processed successfully events.
* Number of delete and add processed events with an error.
* Duration of delete and add processed successfully events.
* Duration of delete and add processed events with an error.

## How to use the recorder with the controller.

When you create a controller you can pass the metrics recorder that you want. 
**Note: If you pass a `nil` backend, it will not record any metric**

If you want a full example, there is a controller [example][metrics-example] that uses the different metric backends

### Prometheus

Prometheus backend needs a prometheus [registerer][prometheus-registerer] and a namespace (a prefix for the metircs). 

For example:

```golang
import (
    ...
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    ...
)

...

    reg := prometheus.NewRegistry()
    m := metrics.NewPrometheus(metricsPrefix, reg)

    ...

    ctrl := controller.NewSequential(30*time.Second, hand, retr, m, log)

    ...

    h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
    logger.Infof("serving metrics at %s", metricsAddr)
    http.ListenAndServe(metricsAddr, h)

    return m
}
```

If you are using the default prometheus methods instead of a custom registry, you could get that from `prometheus.DefaultRegisterer` instead of creating a new registry `reg := prometheus.NewRegistry()`


[metrics-interface]: https://github.com/spotahome/kooper/blob/master/monitoring/metrics/metrics.go
[metrics-example]: https://github.com/spotahome/kooper/tree/master/examples/metrics-controller
[prometheus-registerer]: https://godoc.org/github.com/prometheus/client_golang/prometheus#Registerer
