#!/bin/bash

set -ex

run_setup() {
  kubectl run pod1 --image nginx:latest -l app=pod1
  kubectl run pod2 --image nginx:latest -l app=pod2
  kubectl apply -f np1.yaml
}

run_cleanup() {
  echo
}

run_query() {
  kubectl get pod pod1 || true
  kubectl get pod pod2 || true
}

case "$1" in
  "setup") run_setup ;;
  "cleanup") run_cleanup ;;
  *) run_query ;;
esac

