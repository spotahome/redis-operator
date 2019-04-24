#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Ugly but needs to be relative if we want to use k8s.io/code-generator
# as it is without touching/sed-ing the code/scripts
CODEGEN_PKG=./../../../../..${GOPATH}/src/k8s.io/code-generator

# Add all groups space separated.
GROUPS_VERSION="redisfailover:v1"

# Only generate deepcopy (runtime object needs) and typed client.
# Typed listers & informers not required for the moment. Used with generic
# custom informer/listerwatchers.
${CODEGEN_PKG}/generate-groups.sh "deepcopy,client" \
  github.com/spotahome/redis-operator/client/k8s \
  github.com/spotahome/redis-operator/api \
  "${GROUPS_VERSION}"
