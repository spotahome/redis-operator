FROM golang:1.12-alpine

WORKDIR /go/src/github.com/spotahome/redis-operator
COPY . .
RUN ./scripts/build.sh

FROM alpine:latest
RUN apk --no-cache add \
  ca-certificates
COPY --from=0 /go/src/github.com/spotahome/redis-operator/bin/linux/redis-operator /usr/local/bin
RUN addgroup -g 1000 rf && \
    adduser -D -u 1000 -G rf rf && \
    chown rf:rf /usr/local/bin/redis-operator
USER rf

ENTRYPOINT ["/usr/local/bin/redis-operator"]
