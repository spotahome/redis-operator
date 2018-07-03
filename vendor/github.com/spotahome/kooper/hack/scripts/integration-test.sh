#!/usr/bin/env sh

set -o errexit
set -o nounset

go test `go list ./... | grep test/integration` -v -tags='integration'