#!/usr/bin/env bash

source "./helpers.bash"

set -e

KUBERNETES_HOSTS=(worker0 worker1 worker2)

for host in ${KUBERNETES_HOSTS[*]}; do
  gcloud compute copy-files helpers.bash 05-2-run-inside-vms-kubernetes-worker.sh ${host}:~/
done
