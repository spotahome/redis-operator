#!/bin/sh

export CGO_ENABLED=0
export GOOS=linux
export GO111MODULE=on

go build -ldflags '-extldflags "-static"' -a -v -o bin/linux/redis-operator ./cmd/redisoperator/
