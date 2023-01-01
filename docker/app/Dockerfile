FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS build
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
RUN chown nobody:nobody /usr/local/bin/redis-operator
USER nobody

ENTRYPOINT ["/usr/local/bin/redis-operator"]
