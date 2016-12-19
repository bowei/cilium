#!/usr/bin/env bash

source "./helpers.bash"

set -e

#wget https://storage.googleapis.com/kubernetes-release/release/${k8s_version}/bin/linux/amd64/kubectl
#chmod +x kubectl
#sudo mv kubectl /usr/local/bin

KUBERNETES_PUBLIC_ADDRESS=$(gcloud compute addresses describe kubernetes \
  --format 'value(address)')

kubectl config set-cluster kubernetes-the-hard-way \
  --certificate-authority=ca.pem \
  --embed-certs=true \
  --server=https://${KUBERNETES_PUBLIC_ADDRESS}:6443

kubectl config set-credentials admin --token chAng3m3

kubectl config set-context default-context \
  --cluster=kubernetes-the-hard-way \
  --user=admin

kubectl get componentstatuses

kubectl get nodes
