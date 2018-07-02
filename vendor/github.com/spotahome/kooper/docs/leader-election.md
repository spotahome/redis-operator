# Leader election

Kooper comes with support for leader election.

A controller can be run with multiple instances in HA and only one will be really running the controller logic (the leader), when the leader stops or loses the leadership another instance will take the leadership and so on.

## Usage

The default controllers don't run in leader election mode:

* `controller.NewSequential`
* `controller.NewConcurrent`

To use the leader election you can use:

* `controller.New`

These method accepts a [`leaderelection.Runner`][leaderelection-src] service that manages the leader election of the controller, if this service is a `nil` object the controller will fallback to a regular controller mode.

The leader election comes with two contructors, one that has safe settings for the lock and one that can be configured.

Lets take and example of creating using the default leader election service.

```golang
import (
    ...
    "github.com/spotahome/kooper/operator/controller/leaderelection"
    ...
)

...

lesvc, err := leaderelection.NewDefault("my-controller", "myControllerNS", k8scli, logger)
ctrl := controller.New(cfg, hand, retr, lesvc, nil, logger)
...
```

Another example customizing the lock would be:

```golang
import (
    ...
    "github.com/spotahome/kooper/operator/controller/leaderelection"
    ...
)

...

rlCfg := &leaderelection.LockConfig{
    LeaseDuration: 20 * time.Second
    RenewDeadline: 12 * time.Second
    RetryPeriod:   3 * time.Second
}
lesvc, err := leaderelection.New("my-controller", "myControllerNS", rlCfg,k8scli, logger)
...
```

## Important notes

### Lock

When using the leader election in a controller, the controller needs the namespace where the controller is running, this is because the lock is made using a configmap (that will be on the namespace where the controller is running). Also because of this, it needs to get, create and update a configmap.

This means that if you are using RBAC, the definition would need at least these permissions:

```yaml
rules:
- apiGroups:
    - ""
    resources:
    - configmaps
    verbs:
    - create
    - get
    - update
```

### Losing the leadership

When one of the leaders looses the leadership the controller will end its execution (Kubernetes eventually should spin up a new instance)

## Full example

For a full example check [this][leaderelection-example]

## Test example in local

You can check how it works locally using docker running N controllers in different containers.

For example `ctrl1` and `ctrl2`:

```bash
docker run --name ctrl1 \
    --network bridge \
    --rm -it \
    -v ${HOME}/.kube:/root/.kube:ro \
    -v `pwd`:/go/src/github.com/spotahome/kooper:ro  \
    golang go run /go/src/github.com/spotahome/kooper/examples/leader-election-controller/main.go
```

```bash
docker run --name ctrl2 \
    --network bridge \
    --rm -it \
    -v ${HOME}/.kube:/root/.kube:ro \
    -v `pwd`:/go/src/github.com/spotahome/kooper:ro  \
    golang go run /go/src/github.com/spotahome/kooper/examples/leader-election-controller/main.go
```

Now you can test disconnecting and connecting them using these commands and checking the results.

* `docker network disconnect bridge ctrl2`
* `docker network disconnect bridge ctrl1`
* `docker network connect bridge ctrl2`
* `docker network connect bridge ctrl1`


[leaderelection-src]: https://github.com/spotahome/kooper/tree/master/operator/controller/leaderelection
[leaderelection-example]: https://github.com/spotahome/kooper/tree/master/examples/leader-election-controller