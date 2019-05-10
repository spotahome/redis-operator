FROM golang:1.12-alpine

ENV CODEGEN_VERSION="1.11.9"

RUN apk --no-cache add \
    bash \
    git \
    g++ \
    openssl

# Code generator stuff
# Check: https://github.com/kubernetes/kubernetes/pull/57656
RUN wget http://github.com/kubernetes/code-generator/archive/kubernetes-${CODEGEN_VERSION}.tar.gz && \
    mkdir -p /go/src/k8s.io/code-generator/ && \
    tar zxvf kubernetes-${CODEGEN_VERSION}.tar.gz --strip 1 -C /go/src/k8s.io/code-generator/ && \
    mkdir -p /go/src/k8s.io/kubernetes/hack/boilerplate/ && \
    touch /go/src/k8s.io/kubernetes/hack/boilerplate/boilerplate.go.txt

# Go dep installation
RUN go get -u github.com/golang/dep/cmd/dep \
    && mkdir -p /go/pkg/dep \
    && chmod 777 /go/pkg/dep

# Mock creator
RUN go get github.com/vektra/mockery/.../

# Create user
ARG uid=1000
ARG gid=1000
RUN addgroup -g $gid rf && \
    adduser -D -u $uid -G rf rf && \
    chown rf:rf -R /go


USER rf
WORKDIR /go/src/github.com/spotahome/redis-operator
