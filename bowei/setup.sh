#!/bin/bash

set -euxo pipefile

case "$1" in
  "setup") run_setup ;;
  "cleanup") run_cleanup ;;
  *) run_query ;;
esac

run_setup() {
  kubectl run pod1 --image nginx:latest -l app=pod1
  kubectl run pod2 --image nginx:latest -l app=pod2
  kubectl apply -f np1.yaml
}

run_cleanup() {
}

run_query() {
  kubectl get pod1
  kubectl get pod2
}
