
# Name of this service/application
SERVICE_NAME := kooper

# Path of the go service inside docker
DOCKER_GO_SERVICE_PATH := /src

# Shell to use for running scripts
SHELL := $(shell which bash)

# Get docker path or an empty string
DOCKER := $(shell command -v docker)

# Get the main unix group for the user running make (to be used by docker-compose later)
GID := $(shell id -g)

# Get the unix user id for the user running make (to be used by docker-compose later)
UID := $(shell id -u)

# cmds
UNIT_TEST_CMD := ./hack/scripts/unit-test.sh
INTEGRATION_TEST_CMD := ./hack/scripts/integration-test.sh 
CI_INTEGRATION_TEST_CMD := ./hack/scripts/integration-test-kind.sh
MOCKS_CMD := ./hack/scripts/mockgen.sh
DOCKER_RUN_CMD := docker run -v ${PWD}:$(DOCKER_GO_SERVICE_PATH) --rm -it $(SERVICE_NAME)
RUN_EXAMPLE_POD_ECHO := go run ./examples/echo-pod-controller/cmd/* --development
RUN_EXAMPLE_POD_ECHO_ONEFILE := go run ./examples/onefile-echo-pod-controller/main.go --development
RUN_EXAMPLE_POD_TERM := go run ./examples/pod-terminator-operator/cmd/* --development
DEPS_CMD := GO111MODULE=on go mod tidy && GO111MODULE=on go mod vendor
K8S_VERSION := "1.15.6"
SET_K8S_DEPS_CMD := GO111MODULE=on go mod edit \
    -require=k8s.io/apiextensions-apiserver@kubernetes-${K8S_VERSION} \
	-require=k8s.io/client-go@kubernetes-${K8S_VERSION} \
	-require=k8s.io/apimachinery@kubernetes-${K8S_VERSION} \
	-require=k8s.io/api@kubernetes-${K8S_VERSION} \
	-require=k8s.io/kubernetes@v${K8S_VERSION} && \
	$(DEPS_CMD)

# environment dirs
DEV_DIR := docker/dev

# The default action of this Makefile is to build the development docker image
.PHONY: default
default: build

# Test if the dependencies we need to run this Makefile are installed
.PHONY: deps-development
deps-development:
ifndef DOCKER
	@echo "Docker is not available. Please install docker"
	@exit 1
endif

# Build the development docker image
.PHONY: build
build:
	docker build -t $(SERVICE_NAME) --build-arg uid=$(UID) --build-arg  gid=$(GID) -f ./docker/dev/Dockerfile .

# Dependency stuff.
.PHONY: set-k8s-deps
set-k8s-deps:
	$(SET_K8S_DEPS_CMD)

.PHONY: deps
deps:
	$(DEPS_CMD)

# Test stuff in dev
.PHONY: unit-test
unit-test: build
	$(DOCKER_RUN_CMD) /bin/sh -c '$(UNIT_TEST_CMD)'
.PHONY: integration-test
integration-test: build
	echo "[WARNING] Requires a kubernetes cluster configured (and running) on your kubeconfig!!"
	$(INTEGRATION_TEST_CMD)
.PHONY: test
test: unit-test

# Test stuff in ci
.PHONY: ci-unit-test
ci-unit-test: 
	$(UNIT_TEST_CMD)
.PHONY: ci-integration-test
ci-integration-test:
	$(CI_INTEGRATION_TEST_CMD)
.PHONY: ci
ci: ci-unit-test ci-integration-test

# Mocks stuff in dev
.PHONY: mocks
mocks: build
	$(DOCKER_RUN_CMD) /bin/sh -c '$(MOCKS_CMD)'

# Run examples.
.PHONY: controller-example
controller-example:
	$(RUN_EXAMPLE_POD_ECHO)
.PHONY: controller-example-onefile
controller-example-onefile:
	$(RUN_EXAMPLE_POD_ECHO_ONEFILE)
.PHONY: operator-example
operator-example:
	$(RUN_EXAMPLE_POD_TERM)
