#!/usr/bin/env bash

source "./helpers.bash"

set -e

kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/gce/daemon-set.yaml

kubectl get daemonset cilium-net-controller

kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/gce/network-policy/admin-policy.json
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/gce/network-policy/lizards-policy-db.json
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/gce/network-policy/lizards-policy-web.json

kubectl get networkpolicy

kubectl get pods

while read line; do
cat <<EOF | kubectl exec -i ${line} --  cilium -D policy import -
{
        "name": "io.cilium",
        "rules": [{
                "coverage": ["reserved:host"],
                "allow": ["reserved:all"]
        }]
}
EOF
done < <(kubectl get pods --output=jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')

random_pod=$(kubectl get pods --output=jsonpath='{.items[0].metadata.name}{"\n"}')

kubectl exec ${random_pod} cilium policy dump
