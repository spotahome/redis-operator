# The following are targers that do not exist in the filesystem as real files and should be always executed by make
.PHONY: default build deps-development docker-build shell run image unit-test test generate go-generate get-deps update-deps testing
VERSION := 0.1.6

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

# Get user ID
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

# environment dirs
DEV_DIR := docker/development
APP_DIR := docker/app

# workdir
WORKDIR := /go/src/github.com/spotahome/redis-operator

# The default action of this Makefile is to build the development docker image
default: build

# Test if the dependencies we need to run this Makefile are installed
deps-development:
ifndef DOCKER
	@echo "Docker is not available. Please install docker"
	@exit 1
endif

# Run the development environment in non-daemonized mode (foreground)
docker-build: deps-development
	docker build \
		--build-arg UID=$(UID) \
		-t $(REPOSITORY)-dev:latest \
		-t $(REPOSITORY)-dev:$(COMMIT) \
		-f $(DEV_DIR)/Dockerfile \
		.

# Run a shell into the development docker image
shell: docker-build
	docker run -ti --rm -v ~/.kube:/.kube:ro -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) -p $(PORT):$(PORT) $(REPOSITORY)-dev /bin/bash

# Build redis-failover executable file
build: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev ./scripts/build.sh

# Run the development environment in the background
run: docker-build
	docker run -ti --rm -v ~/.kube:/.kube:ro -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) -p $(PORT):$(PORT) $(REPOSITORY)-dev ./scripts/run.sh

# Build the production image based on the public one
image: deps-development
	docker build \
	-t $(SERVICE_NAME) \
	-t $(REPOSITORY):latest \
	-t $(REPOSITORY):$(COMMIT) \
	-t $(REPOSITORY):$(BRANCH) \
	-f $(APP_DIR)/Dockerfile \
	.

testing: image
	docker push $(REPOSITORY):$(BRANCH)

tag:
	git tag $(VERSION)

publish:
	@COMMIT_VERSION="$$(git rev-list -n 1 $(VERSION))"; \
	docker tag $(REPOSITORY):"$$COMMIT_VERSION" $(REPOSITORY):$(VERSION)
	docker push $(REPOSITORY):$(VERSION)
	docker push $(REPOSITORY):latest

release: tag image publish

# Test stuff in dev
unit-test: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev /bin/sh -c '$(UNIT_TEST_CMD)'
test: unit-test

go-generate: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev /bin/sh -c '$(GO_GENERATE_CMD)'

generate: go-generate

get-deps: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev /bin/sh -c '$(GET_DEPS_CMD)'

update-deps: docker-build
	docker run -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(UID) --name $(SERVICE_NAME) $(REPOSITORY)-dev /bin/sh -c '$(UPDATE_DEPS_CMD)'

# Custom commands
#...
