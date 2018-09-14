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
In order to create Redis failovers inside a Kubernetes cluster, the operator has to be deployed. It can be done with a [deployment](example/operator.yaml) or with the provided [Helm chart](charts/redisoperator).

### Using a Deployment
To create the operator, you can directly create it with kubectl:
```
kubectl create -f https://raw.githubusercontent.com/spotahome/redis-operator/master/example/operator.yaml
```
This will create a deployment named `redisoperator`.

### Using the Helm chart
From the root folder of the project, execute the following:
```
helm install --name redisfailover charts/redisoperator
```

## Usage
Once the operator is deployed inside a Kubernetes cluster, a new API will be accesible, so you'll be able to create, update and delete redisfailovers.

In order to deploy a new redis-failover a [specification](example/redisfailover/all-options.yaml) has to be created:
```
kubectl create -f https://raw.githubusercontent.com/spotahome/redis-operator/master/redisfailover/all-options.yaml
```

This redis-failover will be managed by the operator, resulting in the following elements created inside Kubernetes:
* `rfr-<NAME>`: Redis configmap
* `rfr-<NAME>`: Redis statefulset
* `rfr-<NAME>`: Redis service (if redis-exporter is enabled)
* `rfs-<NAME>`: Sentinel configmap
* `rfs-<NAME>`: Sentinel deployment
* `rfs-<NAME>`: Sentinel service

**NOTE**: `NAME` is the named provided when creating the RedisFailover.
**IMPORTANT**: the name of the redis-failover to be created cannot be longer that 48 characters, due to prepend of redis/sentinel identification and statefulset limitation.

### Persistance
The operator has the ability of add persistance to Redis data. By default an `emptyDir` will be used, so the data is not saved.

In order to have persistance, a PersistentVolumeClaim usage is allowed. The full [PVC definition has to be added](example/redisfailover/persistant-storage.yaml) to the Redis Failover Spec under the `Storage` section.

**IMPORTANT**: By default, the persistent volume claims will be deleted when the Redis Failover is. If this is not the expected usage, a `keepAfterDeletion` flag can be added under the `storage` section of Redis. [An example is given](example/redisfailover/persistant-storage-no-pvc-deletion.yaml).

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

**Important*: this options will be set via `config set` or `sentinel set mymaster`. In the Sentinel options, there are some "conversions" to be made.

- Configuration on the `sentinel.conf`: `sentinel down-after-milliseconds mymaster 2000`
- Configuration on the `configOptions`: `down-after-milliseconds 2000`

**Important 2**: do **NOT** change the options used for control the redis/sentinel such as `port`, `bind`, `dir`, etc.

### Connection
In order to connect to the redis-failover and use it, a [Sentinel-ready](https://redis.io/topics/sentinel-clients) library has to be used. This will connect through the Sentinel service to the Redis node working as a master.
The connection parameters are the following:
```
url: rfs-<NAME>
port: 26379
master-name: mymaster
```

## Cleanup
If you want to delete the operator from your Kubernetes cluster, the operator deployment should be deleted.

Also, the CRD has to be deleted too:
```
kubectl delete crd redisfailovers.storage.spotahome.com
```

## Documentation
For the code documentation, you can lookup on the [GoDoc](https://godoc.org/github.com/spotahome/redis-operator).

Also, you can check more deeply information on the [docs folder](docs).
