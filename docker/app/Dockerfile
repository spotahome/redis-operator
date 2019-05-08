FROM golang:1.12-alpine

WORKDIR /go/src/github.com/spotahome/redis-operator
COPY . .
RUN ./scripts/build.sh

FROM alpine:latest
RUN apk --no-cache add \
  ca-certificates
COPY --from=0 /go/src/github.com/spotahome/redis-operator/bin/linux/redis-operator /usr/local/bin

ENTRYPOINT ["/usr/local/bin/redis-operator"]
