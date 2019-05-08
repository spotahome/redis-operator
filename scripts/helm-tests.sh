#!/bin/bash

set -eu

chart=charts/redisoperator

echo ">> Testing chart ${chart}"

helm lint ${chart}
helm template ${chart}

echo "> Chart OK"
