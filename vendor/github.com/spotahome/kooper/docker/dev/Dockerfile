FROM golang:1.13-alpine

RUN apk --no-cache add \
    g++ \
    git

# Mock creator
RUN go get -u github.com/vektra/mockery/.../

# Create user
ARG uid=1000
ARG gid=1000
RUN addgroup -g $gid kooper && \
    adduser -D -u $uid -G kooper kooper


# Fill go mod cache.
RUN mkdir /tmp/cache
COPY go.mod /tmp/cache
COPY go.sum /tmp/cache
RUN cd /tmp/cache && \
    go mod download

RUN chown kooper:kooper -R /go
USER kooper

WORKDIR /src