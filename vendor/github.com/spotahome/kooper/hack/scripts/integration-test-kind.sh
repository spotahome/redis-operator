#!/bin/bash

set -o errexit
set -o nounset

KUBERNETES_VERSION=v${KUBERNETES_VERSION:-1.15.6}
current_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PREVIOUS_KUBECTL_CONTEXT=$(kubectl config current-context) || PREVIOUS_KUBECTL_CONTEXT=""

function cleanup {
    if [ ! -z $PREVIOUS_KUBECTL_CONTEXT ]
    then
      kubectl config use-context $PREVIOUS_KUBECTL_CONTEXT
    fi
    echo "=> Removing kind cluster"
    kind delete cluster
}
trap cleanup EXIT

echo "=> Preparing kind for running integration tests"
kind create cluster --image kindest/node:${KUBERNETES_VERSION}

kubectl config use-context kind-kind

echo "=> Running integration tests"
${current_dir}/integration-test.sh
