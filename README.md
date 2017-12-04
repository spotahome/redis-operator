# redis-operator [![Build Status](https://travis-ci.org/spotahome/redis-operator.png)](https://travis-ci.org/spotahome/redis-operator) [![Go Report Card](http://goreportcard.com/badge/spotahome/redis-operator)](http://goreportcard.com/report/spotahome/redis-operator)
Redis Operator creates/configures/manages redis clusters atop Kubernetes.

## Requirements
Redis Operator is meant to be run on Kubernetes 1.8+.
All dependecies have been vendored, so there's no need to any additional download.

### Images
#### Redis Operator
[![Redis Operator Image](https://quay.io/repository/spotahome/redis-operator/status "Redis Operator Image")](https://quay.io/repository/spotahome/redis-operator)

#### Redis Operator Toolkit
[![Redis Operator Toolkit Image](https://quay.io/repository/spotahome/redis-operator-toolkit/status "Redis Operator Toolkit Image")](https://quay.io/repository/spotahome/redis-operator-toolkit)

## Operator deployment on kubernetes
In order to create Redis failovers inside a Kubernetes cluster, the operator has to be deployed:
~~~~
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: redisoperator
  name: redisoperator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redisoperator
  strategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: redisoperator
    spec:
      containers:
      - image: quay.io/spotahome/redis-operator:0.1.0
        imagePullPolicy: IfNotPresent
        name: app
        resources:
          limits:
            cpu: 100m
            memory: 50Mi
          requests:
            cpu: 10m
            memory: 50Mi
      restartPolicy: Always
~~~~

## Usage
In order to deploy a new redis-failover inside kubernetes, a specification has to be created. Here is a template:
~~~~
apiVersion: spotahome.com/v1alpha1
kind: RedisFailover
metadata:
  name: myredisfailover
  namespace: mynamespace
spec:
  sentinel:
    replicas: 3        # Optional. Value by default, can be set higher.
    resources:         # Optional. If not set, it won't be defined on created reosurces
      requests:
        cpu: 100m
      limits:
        memory: 100Mi
  redis:
    replicas: 3        # Optional. Value by default, can be set higher.
    resources:         # Optional. If not set, it won't be defined on created reosurces
      requests:
        cpu: 100m
      limits:
        memory: 100Mi
    exporter: false    # Optional. False by default. Adds a redis-exporter container to export metrics.
~~~~

## Creation pipeline
The redis-operator creates a redis failover, using the following pipeline:

1. Start a redis bootstrap pod, containing a redis as a master allowing other redis/sentinel to connect to it.
2. Create a sentinel service to allow service discovery.
3. Start a sentinel deployment with the number of replicas set by the user definition (or 3 by default). This replicas monitor the bootstrap master.
4. Create a pod disruption budget for sentinel pods.
5. If the redis exporter is active, create a headless redis service for discovery.
6. Start a redis statefulset, who connects to the redis bootstrap pod as a slaves.
7. Create a pod disruption budget for redis pods.
8. Delete the redis bootstrap pod.
9. From this moment, the redis-failover will check the cluster is ok. If not, it will try to fix it when possible. Normal failures will be controlled by sentinel.

## Code folder structure
* cmd: contains the starting point of the application.
* mocks: contains the mocked interfaces for testing the application.
* pkg:
  * clock: wrapper of time, created to be able to mock it.
  * config: contains the constants of the application.
  * failover: contains the logic of the application.
  * log: wrapper of logrus, created to be able to mock it.
  * redis: interface wich allows separate the library used and the logic.
  * tpr: created to define the third party resource that will be registered into k8s.
* vendor: vendored packages used by the application.

## No code folder structure
* charts: helm chart to deploy the TPR.
* docker: Dockerfiles to generate redis-failover docker images.
* example: yaml files with spec of redis-failover.
* scripts: scripts used to build and run the app.

## Development
### With Make
You can do the following commands with make:
* Build the development container.
`make docker-build`
* Generate mocks.
`make go-generate`
* Run tests.
`make test`
* Build the executable file.
`make build`
* Run the app.
`make run`
* Access the docker instance with a shell.
`make shell`
* Install dependencies
`make get-deps`
* Update dependencies
`make update-deps`
* Build the app image.
`make image`

