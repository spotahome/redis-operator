#!/usr/bin/env bash

set -o errexit
set -o nounset

src=./cmd/redisoperator
out=./bin/redis-operator

final_out=${out}
ldf_cmp="-w -extldflags '-static'"
f_ver="-X main.Version=${VERSION:-dev}"

echo "Building binary at ${out}"
CGO_ENABLED=0 go build -o ${out} --ldflags "${ldf_cmp} ${f_ver}"  ${src}
