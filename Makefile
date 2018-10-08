VERSION := 0.5.4

# Name of this service/application
SERVICE_NAME := redis-operator

# Docker image name for this project
IMAGE_NAME := spotahome/$(SERVICE_NAME)

# Repository url for this project
REPOSITORY := quay.io/$(IMAGE_NAME)

# Shell to use for running scripts
SHELL := $(shell which bash)

# Get docker path or an empty string
DOCKER := $(shell command -v docker)

# Get the main unix group for the user running make (to be used by docker-compose later)
GID := $(shell id -g)

# Get the unix user id for the user running make (to be used by docker-compose later)
UID := $(shell id -u)

# Commit hash from git
COMMIT=$(shell git rev-parse HEAD)

# Branch from git
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

PORT := 9710

# CMDs
UNIT_TEST_CMD := go test `go list ./... | grep -v /vendor/` -v
GO_GENERATE_CMD := go generate `go list ./... | grep -v /vendor/`
GET_DEPS_CMD := dep ensure
UPDATE_DEPS_CMD := dep ensure
UPDATE_CODEGEN_CMD := ./hack/update-codegen.sh
MOCKS_CMD := go generate ./mocks

# environment dirs
DEV_DIR := docker/development
APP_DIR := docker/app

# workdir
WORKDIR := /go/src/github.com/spotahome/redis-operator

# The default action of this Makefile is to build the development docker image
.PHONY: default
default: build

# Run the development environment in non-daemonized mode (foreground)
.PHONY: docker-build
docker-build: deps-development
	docker build \
		--build-arg UID=$(UID) \
		-t $(REPOSITORY)-dev:latest \
		-t $(REPOSITORY)-dev:$(COMMIT) \
		-f $(DEV_DIR)/Dockerfile \
		.

# Run a shell into the development docker image
.PHONY: shell
shell: docker-build
	docker run -ti --rm -v ~/.kube:/.kube:ro -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) -p $(PORT):$(PORT) $(REPOSITORY)-dev /bin/bash

# Build redis-failover executable file
.PHONY: build
build: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev ./scripts/build.sh

# Run the development environment in the background
.PHONY: run
run: docker-build
	docker run -ti --rm -v ~/.kube:/.kube:ro -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) -p $(PORT):$(PORT) $(REPOSITORY)-dev ./scripts/run.sh

# Build the production image based on the public one
.PHONY: image
image: deps-development
	docker build \
	-t $(SERVICE_NAME) \
	-t $(REPOSITORY):latest \
	-t $(REPOSITORY):$(COMMIT) \
	-t $(REPOSITORY):$(BRANCH) \
	-f $(APP_DIR)/Dockerfile \
	.

.PHONY: testing
testing: image
	docker push $(REPOSITORY):$(BRANCH)

.PHONY: tag
tag:
	git tag $(VERSION)

.PHONY: publish
publish:
	@COMMIT_VERSION="$$(git rev-list -n 1 $(VERSION))"; \
	docker tag $(REPOSITORY):"$$COMMIT_VERSION" $(REPOSITORY):$(VERSION)
	docker push $(REPOSITORY):$(VERSION)
	docker push $(REPOSITORY):latest

.PHONY: release
release: tag image publish

# Test stuff in dev
.PHONY: unit-test
unit-test: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev /bin/sh -c '$(UNIT_TEST_CMD)'

# Run both integration and unit tests
.PHONY: integration-test
integration-test:
	./scripts/integration-tests.sh

.PHONY: test
test: unit-test integration-test

.PHONY: go-generate
go-generate: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev /bin/sh -c '$(GO_GENERATE_CMD)'

.PHONY: generate
generate: go-generate

.PHONY: get-deps
get-deps: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev /bin/sh -c '$(GET_DEPS_CMD)'

.PHONY: update-deps
update-deps: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev /bin/sh -c '$(UPDATE_DEPS_CMD)'

.PHONY: mocks
mocks: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev /bin/sh -c '$(MOCKS_CMD)'

.PHONY: deps-development
# Test if the dependencies we need to run this Makefile are installed
deps-development:
ifndef DOCKER
	@echo "Docker is not available. Please install docker"
	@exit 1
endif

# Generate kubernetes code for types..
.PHONY: update-codegen
update-codegen: docker-build
	@echo ">> Generating code for Kubernetes CRD types..."
	docker run --rm -v $(PWD):/go/src/github.com/spotahome/redis-operator/ $(REPOSITORY)-dev /bin/bash -c '$(UPDATE_CODEGEN_CMD)'
