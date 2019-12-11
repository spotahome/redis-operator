#!/bin/bash

set -eu

SUDO=''
if [[ $(id -u) -ne 0 ]]
then
    SUDO="sudo"
fi

function cleanup {
    echo "=> Removing minikube cluster"
    $SUDO minikube delete
}
trap cleanup EXIT

echo "=> Preparing minikube for running integration tests"
$SUDO minikube start --vm-driver=none --kubernetes-version=v1.15.6

echo "=> Waiting for minikube to start"
sleep 30

# Hack for Travis. The kubeconfig has to be readable
if [[ -v IN_TRAVIS ]]
then
    $SUDO chown -R travis: ${HOME}/.minikube/
    $SUDO chmod a+r ${HOME}/.kube/config
fi

echo "=> Running integration tests"
go test `go list ./... | grep test/integration` -v -tags='integration'
