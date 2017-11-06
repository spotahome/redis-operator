#!/bin/sh

./scripts/build.sh && ./bin/linux/redis-operator --kubeconfig=/.kube/config
