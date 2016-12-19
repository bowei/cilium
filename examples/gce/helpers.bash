#!/usr/bin/env bash

region="europe-west1"
zone="${region}-b"

k8s_cluster_range="172.30.0.0/24"

k8s_service_cluster_range="172.26.0.0/16"
k8s_master_service_ip="172.26.0.1"
cluster_dns_ip="172.26.0.10"

k8s_version="v1.5.1"
etcd_version="v3.1.0-rc.1"

docker_version="1.12.5"

# Not advised to change the number of controllers and number of worker unless you see all
# scripts, they are hardcoded for 3 controllers and 3 workers.
controllers_ips=("172.30.0.10" "172.30.0.11" "172.30.0.12")
workers_ips=("172.30.0.20" "172.30.0.21" "172.30.0.22")