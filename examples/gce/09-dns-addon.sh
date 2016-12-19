#!/usr/bin/env bash

source "./helpers.bash"

set -e

kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/gce/kubedns-svc.yaml

kubectl --namespace=kube-system get svc

kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/gce/kubedns-rc.yaml

kubectl --namespace=kube-system get pods
