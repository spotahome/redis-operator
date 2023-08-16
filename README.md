# redis-operator

[![Build Status](https://github.com/spotahome/redis-operator/actions/workflows/ci.yaml/badge.svg?branch=master)](https://github.com/spotahome/redis-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/spotahome/redis-operator)](https://goreportcard.com/report/github.com/spotahome/redis-operator)

Redis Operator creates/configures/manages redis-failovers atop Kubernetes.

## Requirements

Kubernetes version: 1.21 or higher
Redis version: 6 or higher

Redis operator is being tested against kubernetes 1.25 1.26 1.27 and redis 6
All dependencies have been vendored, so there's no need to any additional download.

## Operator deployment on Kubernetes

In order to create Redis failovers inside a Kubernetes cluster, the operator has to be deployed.
It can be done with plain old [deployment](example/operator), using [Kustomize](manifests/kustomize) or with the provided [Helm chart](charts/redisoperator).

### Using the Helm chart

From the root folder of the project, execute the following:

```
helm repo add redis-operator https://spotahome.github.io/redis-operator
helm repo update
helm install redis-operator redis-operator/redis-operator
```

#### Update helm chart

Helm chart only manage the creation of CRD in the first install. In order to update the CRD you will need to apply directly.

```
REDIS_OPERATOR_VERSION=v1.3.0
kubectl replace -f https://raw.githubusercontent.com/spotahome/redis-operator/${REDIS_OPERATOR_VERSION}/manifests/databases.spotahome.com_redisfailovers.yaml
```

```
helm upgrade redis-operator redis-operator/redis-operator
```
### Using kubectl

To create the operator, you can directly create it with kubectl:

```
REDIS_OPERATOR_VERSION=v1.3.0
kubectl create -f https://raw.githubusercontent.com/spotahome/redis-operator/${REDIS_OPERATOR_VERSION}/manifests/databases.spotahome.com_redisfailovers.yaml
kubectl apply -f https://raw.githubusercontent.com/spotahome/redis-operator/${REDIS_OPERATOR_VERSION}/example/operator/all-redis-operator-resources.yaml
```

This will create a deployment named `redisoperator`.

### Using kustomize

The kustomize setup included in this repo is highly customizable using [components](https://kubectl.docs.kubernetes.io/guides/config_management/components/),
but it also comes with a few presets (in the form of overlays) supporting the most common use cases.

To install the operator with default settings and every necessary resource (including RBAC, service account, default resource limits, etc), install the `default` overlay:

```shell
kustomize build github.com/spotahome/redis-operator/manifests/kustomize/overlays/default
```

If you would like to customize RBAC or the service account used, you can install the `minimal` overlay.

Finally, you can install the `full` overlay if you want everything this operator has to offer, including Prometheus ServiceMonitor resources.

It's always a good practice to pin the version of the operator in your configuration to make sure you are not surprised by changes on the latest development branch:

```shell
kustomize build github.com/spotahome/redis-operator/manifests/kustomize/overlays/default?ref=v1.2.4
```

You can easily create your own config by creating a `kustomization.yaml` file
(for example to apply custom resource limits, to add custom labels or to customize the namespace):

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: redis-operator

commonLabels:
    foo: bar

resources:
  - github.com/spotahome/redis-operator/manifests/kustomize/overlays/full
```

Take a look at the manifests inside [manifests/kustomize](manifests/kustomize) for more details.

## Usage

Once the operator is deployed inside a Kubernetes cluster, a new API will be accesible, so you'll be able to create, update and delete redisfailovers.

In order to deploy a new redis-failover a [specification](example/redisfailover/basic.yaml) has to be created:

```
REDIS_OPERATOR_VERSION=v1.2.4
kubectl create -f https://raw.githubusercontent.com/spotahome/redis-operator/${REDIS_OPERATOR_VERSION}/example/redisfailover/basic.yaml
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

## Topology Spread Contraints

You can use the `topologySpreadContraints` to ensure the pods of a type(redis or sentinel) are evenly distributed across zones/nodes. Examples are for using [topology spread constraints](example/redisfailover/topology-spread-contraints.yaml). Further document on how `topologySpreadConstraints` work could be found [here](https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/).

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
`redisfailover` object. See the [SecurityContext example file](example/redisfailover/security-context.yaml) for an example. You can visit kubernetes documentation for detailed docs about [security context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

### Custom containerSecurityContext at container level

By default Kubernetes will run containers with default docker capabilities for exemple, this is not always desirable.
If you need the containers to run with specific capabilities or with read only root file system (or provide any other securityContext options) then you can specify a custom `containerSecurityContext` in the
`redisfailover` object. See the [ContainerSecurityContext example file](example/redisfailover/container-security-context.yaml) for an example. Keys available under containerSecurityContext are detailed [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#securitycontext-v1-core)

### Custom command

By default, redis and sentinel will be called with the basic command, giving the configuration file:

- Redis: `redis-server /redis/redis.conf`
- Sentinel: `redis-server /redis/sentinel.conf --sentinel`

If necessary, this command can be changed with the `command` option inside redis/sentinel spec. An example can be found in the [custom command example file](example/redisfailover/custom-command.yaml).

### Custom Priority Class
In order to use a custom Kubernetes [Priority Class](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass) for Redis and/or Sentinel pods, you can set the `priorityClassName` in the redis/sentinel spec, this attribute has no default and depends on the specific cluster configuration. **Note:** the operator doesn't create the referenced `Priority Class` resource.

### Custom Service Account
In order to use a custom Kubernetes [Service Account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/) for Redis and/or Sentinel pods, you can set the `serviceAccountName` in the redis/sentinel spec, if not specified the `default` Service Account will be used. **Note:** the operator doesn't create the referenced `Service Account` resource.

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


### ExtraVolumes and ExtraVolumeMounts

If the user choose to have extra volumes creates and mounted, he could use the `extraVolumes` and `extraVolumeMounts`, in `spec.redis` of the CRD. This allows users to mount the extra configurations, or secrets to be used. A typical use case for this might be
- Secrets that sidecars might use to backup of RDBs
- Extra users and their secrets and acls that could used the initContainers to create multiple users
- Extra Configurations that could merge on top the existing configurations
- To pass failover scripts for addition for additional operations

```
---
apiVersion: v1
kind: Secret
metadata:
  name: foo
  namespace: exm
type: Opaque
stringData:
  password: MWYyZDFlMmU2N2Rm
---
apiVersion: databases.spotahome.com/v1
kind: RedisFailover
metadata:
  name: foo
  namespace: exm
spec:
  sentinel:
    replicas: 3
    extraVolumes:
    - name: foo
      secret:
        secretName: foo
        optional: false
    extraVolumeMounts:
    - name: foo
      mountPath: "/etc/foo"
      readOnly: true
  redis:
    replicas: 3
    extraVolumes:
    - name: foo
      secret:
        secretName: foo
        optional: false
    extraVolumeMounts:
    - name: foo
      mountPath: "/etc/foo"
      readOnly: true
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

### Bootstrapping from pre-existing Redis Instance(s)
If you are wanting to migrate off of a pre-existing Redis instance, you can provide a `bootstrapNode` to your `RedisFailover` resource spec.

This `bootstrapNode` can be configured as follows:
|       Key      | Type         | Description                                                                                                                                                                               | Example File                                                                                 |
|:--------------:|--------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------|
| host           | **required** | The IP of the target Redis address or the ClusterIP of a pre-existing Kubernetes Service targeting Redis pods                                                                             | [bootstrapping.yaml](example/redisfailover/bootstrapping.yaml)                               |
| port           | _optional_   | The Port that the target Redis address is listening to. Defaults to `6379`.                                                                                                               | [bootstrapping-with-port.yaml](example/redisfailover/bootstrapping-with-port.yaml)           |
| allowSentinels | _optional_   | Allow the Operator to also create the specified Sentinel resources and point them to the target Node/Port. By default, the Sentinel resources will **not** be created when bootstrapping. | [bootstrapping-with-sentinels.yaml](example/redisfailover/bootstrapping-with-sentinels.yaml) |

#### What is Bootstrapping?
When a `bootstrapNode` is provided, the Operator will always set all of the defined Redis instances to replicate from the provided `bootstrapNode` host value.
This allows for defining a `RedisFailover` that replicates from an existing Redis instance to ease cutover from one instance to another.

**Note: Redis instance will always be configured with `replica-priority 0`. This means that these Redis instances can _never_ be promoted to a `master`.**

Depending on the configuration provided, the Operator will launch the `RedisFailover` in two bootstrapping states: without sentinels and with sentinels.

#### Default Bootstrapping Mode (Without Sentinels)
By default, if the `RedisFailover` resource defines a valid `bootstrapNode`, **only the redis instances will be created**.
This allows for ease of bootstrapping from an existing `RedisFailover` instance without the Sentinels intermingling with each other.

#### Bootstrapping With Sentinels
When `allowSentinels` is provided, the Operator will also create the defined Sentinel resources. These sentinels will be configured to point to the provided
`bootstrapNode` as their monitored master.

### Default versions

The image versions deployed by the operator can be found on the [defaults file](api/redisfailover/v1/defaults.go).
## Cleanup

### Operator and CRD

If you want to delete the operator from your Kubernetes cluster, the operator deployment should be deleted.

Also, the CRD has to be deleted. Deleting CRD automatically wil delete all redis failover custom resources and their managed resources:

```
kubectl delete crd redisfailovers.databases.spotahome.com
```

### Single Redis Failover

Thanks to Kubernetes' `OwnerReference`, all the objects created from a redis-failover will be deleted after the custom resource is.

```
kubectl delete redisfailover <NAME>
```

## Docker Images

### Redis Operator

[![Redis Operator Image](https://quay.io/repository/spotahome/redis-operator/status "Redis Operator Image")](https://quay.io/repository/spotahome/redis-operator)
## Documentation

For the code documentation, you can lookup on the [GoDoc](https://godoc.org/github.com/spotahome/redis-operator).

Also, you can check more deeply information on the [docs folder](docs).
