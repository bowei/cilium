#!/usr/bin/env bash

source "./helpers.bash"

set -e

KUBERNETES_HOSTS=(controller0 controller1 controller2)

for host in ${KUBERNETES_HOSTS[*]}; do
  gcloud compute copy-files 04-2-run-inside-vms-kubernetes-controller.sh ${host}:~/
done
