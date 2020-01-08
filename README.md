# redis-operator

[![Build Status](https://travis-ci.org/spotahome/redis-operator.png)](https://travis-ci.org/spotahome/redis-operator)
[![Go Report Card](http://goreportcard.com/badge/spotahome/redis-operator)](http://goreportcard.com/report/spotahome/redis-operator)

Redis Operator creates/configures/manages redis-failovers atop Kubernetes.

## Requirements

Redis Operator is meant to be run on Kubernetes 1.9+.
All dependencies have been vendored, so there's no need to any additional download.

### Versions deployed

The image versions deployed by the operator can be found on the [defaults file](api/redisfailover/v1/defaults.go).

## Images

### Redis Operator

[![Redis Operator Image](https://quay.io/repository/spotahome/redis-operator/status "Redis Operator Image")](https://quay.io/repository/spotahome/redis-operator)

## Operator deployment on kubernetes

In order to create Redis failovers inside a Kubernetes cluster, the operator has to be deployed. It can be done with [deployment](example/operator) or with the provided [Helm chart](charts/redisoperator).

### Using a Deployment

To create the operator, you can directly create it with kubectl:

```
kubectl create -f https://raw.githubusercontent.com/spotahome/redis-operator/master/example/operator/all-redis-operator-resources.yaml
```

This will create a deployment named `redisoperator`.

### Using the Helm chart

From the root folder of the project, execute the following:

```
helm install --name redisfailover charts/redisoperator
```

## Usage

Once the operator is deployed inside a Kubernetes cluster, a new API will be accesible, so you'll be able to create, update and delete redisfailovers.

In order to deploy a new redis-failover a [specification](example/redisfailover/basic.yaml) has to be created:

```
kubectl create -f https://raw.githubusercontent.com/spotahome/redis-operator/master/example/redisfailover/basic.yaml
```

This redis-failover will be managed by the operator, resulting in the following elements created inside Kubernetes:

- `rfr-<NAME>`: Redis configmap
- `rfr-<NAME>`: Redis statefulset
- `rfr-<NAME>`: Redis service (if redis-exporter is enabled)
- `rfs-<NAME>`: Sentinel configmap
- `rfs-<NAME>`: Sentinel deployment
- `rfs-<NAME>`: Sentinel service

**NOTE**: `NAME` is the named provided when creating the RedisFailover.
**IMPORTANT**: the name of the redis-failover to be created cannot be longer that 48 characters, due to prepend of redis/sentinel identification and statefulset limitation.

### Persistence

The operator has the ability of add persistence to Redis data. By default an `emptyDir` will be used, so the data is not saved.

In order to have persistence, a `PersistentVolumeClaim` usage is allowed. The full [PVC definition has to be added](example/redisfailover/persistent-storage.yaml) to the Redis Failover Spec under the `Storage` section.

**IMPORTANT**: By default, the persistent volume claims will be deleted when the Redis Failover is. If this is not the expected usage, a `keepAfterDeletion` flag can be added under the `storage` section of Redis. [An example is given](example/redisfailover/persistent-storage-no-pvc-deletion.yaml).

### NodeAffinity and Tolerations

You can use NodeAffinity and Tolerations to deploy Pods to isolated groups of Nodes. Examples are given for [node affinity](example/redisfailover/node-affinity.yaml), [pod anti affinity](example/redisfailover/pod-anti-affinity.yaml) and [tolerations](example/redisfailover/tolerations.yaml).

### Custom configurations

It is possible to configure both Redis and Sentinel. This is done with the `customConfig` option inside their spec. It is a list of configurations and their values. Example are given in the [custom config example file](example/redisfailover/custom-config.yaml).

In order to have the ability of this configurations to be changed "on the fly", without the need of reload the redis/sentinel processes, the operator will apply them with calls to the redises/sentinels, using `config set` or `sentinel set mymaster` respectively. Because of this, **no changes on the configmaps** will appear regarding this custom configurations and the entries of `customConfig` from Redis spec will not be written on `redis.conf` file. To verify the actual Redis configuration use [`redis-cli CONFIG GET *`](https://redis.io/commands/config-get).

**Important**: in the Sentinel options, there are some "conversions" to be made:

- Configuration on the `sentinel.conf`: `sentinel down-after-milliseconds mymaster 2000`
- Configuration on the `configOptions`: `down-after-milliseconds 2000`

**Important 2**: do **NOT** change the options used for control the redis/sentinel such as `port`, `bind`, `dir`, etc.

### Custom shutdown script

By default, a custom shutdown file is given. This file makes redis to `SAVE` it's data, and in the case that redis is master, it'll call sentinel to ask for a failover.

This behavior is configurable, creating a configmap and indicating to use it. An example about how to use this option can be found on the [shutdown example file](example/redisfailover/custom-shutdown.yaml).

**Important**: the configmap has to be in the same namespace. The configmap has to have a `shutdown.sh` data, containing the script.

### Custom SecurityContext

By default Kubernetes will run containers as the user specified in the Dockerfile (or the root user if not specified), this is not always desirable.
If you need the containers to run as a specific user (or provide any other PodSecurityContext options) then you can specify a custom `securityContext` in the
`redisfailover` object. See the [SecurityContext example file](example/redisfailover/security-context.yaml) for an example. Keys available under securityContext are detailed [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#podsecuritycontext-v1-core)

### Custom command

By default, redis and sentinel will be called with de basic command, giving the configuration file:

- Redis: `redis-server /redis/redis.conf`
- Sentinel: `redis-server /redis/sentinel.conf --sentinel`

If necessary, this command can be changed with the `command` option inside redis/sentinel spec. An example can be found in the [custom command example file](example/redisfailover/custom-command.yaml).

### Custom Pod Annotations
By default, no pod annotations will be applied to Redis nor Sentinel pods.

In order to apply custom pod Annotations, you can provide the `podAnnotations` option inside redis/sentinel spec. An example can be found in the [custom annotations example file](example/redisfailover/custom-annotations.yaml).

### Custom Service Annotations
By default, no service annotations will be applied to the Redis nor Sentinel services.

In order to apply custom service Annotations, you can provide the `serviceAnnotations` option inside redis/sentinel spec. An example can be found in the [custom annotations example file](example/redisfailover/custom-annotations.yaml).

### Control of label propagation.
By default the operator will propagate all labels on the CRD down to the resources that it creates.  This can be problematic if the
labels on the CRD are not fully under your own control (for example: being deployed by a gitops operator)
as a change to a labels value can fail on immutable resources such as PodDisruptionBudgets.  To control what labels the operator propagates
to resource is creates you can modify the labelWhitelist option in the spec.

By default specifying no whitelist or an empty whitelist will cause all labels to still be copied as not to break backwards compatibility.

Items in the array should be regular expressions, see [here](example/redisfailover/control-label-propagation.yaml) as an example of how they can be used and
[here](https://github.com/google/re2/wiki/Syntax) for a syntax reference.

The whitelist can also be used as a form of blacklist by specifying a regular expression that will not match any label.

NOTE: The operator will always add the labels it requires for operation to resources.  These are the following:
```
app.kubernetes.io/component
app.kubernetes.io/managed-by
app.kubernetes.io/name
app.kubernetes.io/part-of
redisfailovers.databases.spotahome.com/name
```

## Connection to the created Redis Failovers

In order to connect to the redis-failover and use it, a [Sentinel-ready](https://redis.io/topics/sentinel-clients) library has to be used. This will connect through the Sentinel service to the Redis node working as a master.
The connection parameters are the following:

```
url: rfs-<NAME>
port: 26379
master-name: mymaster
```

### Enabling redis auth

To enable auth create a secret with a password field:

```
echo -n "pass" > password
kubectl create secret generic redis-auth --from-file=password

## example config
apiVersion: databases.spotahome.com/v1
kind: RedisFailover
metadata:
  name: redisfailover
spec:
  sentinel:
    replicas: 3
  redis:
    replicas: 1
  auth:
    secretPath: redis-auth
```
You need to set secretPath as the secret name which is created before.

## Cleanup

### Operator and CRD

If you want to delete the operator from your Kubernetes cluster, the operator deployment should be deleted.

Also, the CRD has to be deleted too:

```
kubectl delete crd redisfailovers.databases.spotahome.com
```

### Single Redis Failover

Thanks to Kubernetes' `OwnerReference`, all the objects created from a redis-failover will be deleted after the custom resource is.

```
kubectl delete redisfailover <NAME>
```

## Documentation

For the code documentation, you can lookup on the [GoDoc](https://godoc.org/github.com/spotahome/redis-operator).

Also, you can check more deeply information on the [docs folder](docs).
