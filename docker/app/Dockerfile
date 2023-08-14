FROM --platform=$BUILDPLATFORM golang:1.20-alpine AS build
RUN apk --no-cache add \
  bash

WORKDIR /src
COPY . .

ARG TARGETOS 
ARG TARGETARCH 
ARG VERSION
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH VERSION=$VERSION ./scripts/build.sh

FROM alpine:latest
RUN apk --no-cache add \
  ca-certificates
COPY --from=build /src/bin/redis-operator /usr/local/bin
RUN addgroup -g 1000 rf && \
  adduser -D -u 1000 -G rf rf && \
  chown rf:rf /usr/local/bin/redis-operator
USER rf

ENTRYPOINT ["/usr/local/bin/redis-operator"]
