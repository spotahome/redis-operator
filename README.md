# redis-operator

[![Build Status](https://travis-ci.org/spotahome/redis-operator.png)](https://travis-ci.org/spotahome/redis-operator)
[![Go Report Card](http://goreportcard.com/badge/spotahome/redis-operator)](http://goreportcard.com/report/spotahome/redis-operator)

**NOTE**: This is an alpha-status project. We do regular tests on the code and functionality, but we can not assure a production-ready stability.

Redis Operator creates/configures/manages redis-failovers atop Kubernetes.

## Requirements

Redis Operator is meant to be run on Kubernetes 1.8+.
All dependecies have been vendored, so there's no need to any additional download.

### Versions deployed

The image versions deployed by the operator can be found on the [constants file](operator/redisfailover/service/constants.go) for the RedisFailover service.

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

You can use NodeAffinity and Tolerations to deploy Pods to isolated groups of Nodes

Example:

```yaml
apiVersion: v1
items:
  - apiVersion: storage.spotahome.com/v1alpha2
    kind: RedisFailover
    metadata:
      name: redis
    spec:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
                - key: kops.k8s.io/instancegroup
                  operator: In
                  values:
                    - productionnodes
      hardAntiAffinity: false
      redis: null
      sentinel:
        replicas: 3
        resources:
          limits:
            memory: 100Mi
          requests:
            cpu: 100m
      tolerations:
        - effect: NoExecute
          key: dedicated
          operator: Equal
          value: production
kind: List
```

### Custom configurations

It is possible to configure both Redis and Sentinel. This is done with the `customConfig` option inside their spec. It is a list of configurations and their values.

Example:

```yaml
sentinel:
  customConfig:
    - "down-after-milliseconds 2000"
    - "failover-timeout 3000"
redis:
  customConfig:
    - "maxclients 100"
    - "hz 50"
```

In order to have the ability of this configurations to be changed "on the fly", without the need of reload the redis/sentinel processes, the operator will apply them with calls to the redises/sentinels, using `config set` or `sentinel set mymaster` respectively. Because of this, **no changes on the configmaps** will appear regarding this custom configurations and the entries of `customConfig` from Redis spec will not be written on `redis.conf` file. To verify the actual Redis configuration use [`redis-cli CONFIG GET *`](https://redis.io/commands/config-get).

**Important**: in the Sentinel options, there are some "conversions" to be made:

- Configuration on the `sentinel.conf`: `sentinel down-after-milliseconds mymaster 2000`
- Configuration on the `configOptions`: `down-after-milliseconds 2000`

**Important 2**: do **NOT** change the options used for control the redis/sentinel such as `port`, `bind`, `dir`, etc.

### Custom shutdown script

By default, a custom shutdown file is given. This file makes redis to `SAVE` it's data, and in the case that redis is master, it'll call sentinel to ask for a failover.

This behavior is configurable, creating a configmap and indicating to use it. An example about how to use this option can be found on the [shutdown example file](example/redisfailover/custom-shutdown.yaml).

**Important**: the configmap has to be in the same namespace. The configmap has to have a `shutdown.sh` data, containing the script.

### Custom command

By default, redis and sentinel will be called with de basic command, giving the configuration file:

- Redis: `redis-server /redis/redis.conf`
- Sentinel: `redis-server /redis/sentinel.conf --sentinel`

If necessary, this command can be changed with the `command` option inside redis/sentinel spec:

```yaml
sentinel:
  command:
    - "redis-server"
    - "/redis/sentinel.conf"
    - "--sentinel"
    - "--protected-mode"
    - "no"
redis:
  command:
    - "redis-server"
    - "/redis/redis.conf"
    - "--protected-mode"
    - "no"
```

### Connection

In order to connect to the redis-failover and use it, a [Sentinel-ready](https://redis.io/topics/sentinel-clients) library has to be used. This will connect through the Sentinel service to the Redis node working as a master.
The connection parameters are the following:

```
url: rfs-<NAME>
port: 26379
master-name: mymaster
```

#### Connection example

- To get Sentinel service's port
  ```
  kubectl get service -l component=sentinel
  NAME           TYPE         CLUSTER-IP     EXTERNAL-IP    PORT(S)       AGE
  rfs-<NAME>     ClusterIP    10.99.222.41   <none>         26379/TCP     20m
  ```
- To get a Sentinel's name
  ```
  kubectl get pods -l component=sentinel
  NAME              READY   STATUS    RESTARTS   AGE
  rfs-<NAME>        1/1     Running   0          20m
  ```
- To get network information of the Redis node working as a master
  ```
  kubectl exec -it rfs-<NAME> -- redis-cli -p 26379 SENTINEL get-master-addr-by-name mymaster
  1) "10.244.2.15"
  2) "6379"
  ```
- To set a `key:value` pair in Redis master
  ```
  kubectl exec -it rfs-<NAME> -- redis-cli -h 10.244.2.15 -p 6379 SET hello world!
  OK
  ```
- To get `value` from `key`
  ```
  kubectl exec -it rfs-<NAME>-- redis-cli -h 10.244.2.15 -p 6379 GET hello
  "world!"
  ```

## Cleanup

### Operator and CRD

If you want to delete the operator from your Kubernetes cluster, the operator deployment should be deleted.

Also, the CRD has to be deleted too:

```
kubectl delete crd redisfailovers.storage.spotahome.com
```

### Single Redis Failover

Thanks to Kubernetes' `OwnerReference`, all the objects created from a redis-failover will be deleted after the custom resource is.

```
kubectl delete redisfailover <NAME>
```

## Documentation

For the code documentation, you can lookup on the [GoDoc](https://godoc.org/github.com/spotahome/redis-operator).

Also, you can check more deeply information on the [docs folder](docs).
