#!/bin/bash

set -e

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
$SUDO minikube start --vm-driver=none

# Wait a few seconds for Kubernetes to run
sleep 30

# Hack for Travis. The kubeconfig has to be readable
if [ -f /home/travis/.kube/config ]
then
    $SUDO chmod a+rw /home/travis/.kube/config
fi

echo "=> Running integration tests"
go test `go list ./... | grep test/integration` -v -tags='integration'