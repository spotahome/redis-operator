FROM golang:1.8-alpine

WORKDIR /go/src/github.com/spotahome/redis-operator

RUN set -eux \
 && apk add --no-cache \
        curl \
        git \
        gcc \
        musl-dev \
	bash \
 && go get -u github.com/golang/dep/cmd/dep \
 && mkdir -p /go/pkg/dep \
 && chmod 777 /go/pkg/dep

# Mock creator
RUN go get github.com/vektra/mockery/.../

ARG UID
RUN adduser -D -S -h /go/src/github.com/spotahome/redis-operator -u ${UID} developer
