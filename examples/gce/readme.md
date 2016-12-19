# Cilium + Kubernetes The Hard Way from Kelsey Hightower

We thought the best way on how we could provide a tutorial to deploy cilium on gce with
Kubernetes. Giving the fact that [Kelsey's tutorial](https://github.com/kelseyhightower/kubernetes-the-hard-way)
is super dope, he decided to modify it a little bit to accomudate cilium. Most of the
steps will be the same as that tutorial but we will highlight the different parts on each
section. (Spoiler alert: kube-proxy won't be used)

## Cluster Details

* Kubernetes v1.5.1
* Docker 1.12.5
* etcd v3.1.0-rc.1 (cilium minimum requirement)
* [CNI Based Networking](https://github.com/containernetworking/cni)
* CNI Plugin: cilium (+ loopback)
* Network Policy Enforcer: cilium
* Secure communication between all components (etcd, control plane, workers)
* Default Service Account and Secrets

## Platforms

This tutorial assumes you have access to one of the following:

* [Google Cloud Platform](https://cloud.google.com) and the [Google Cloud SDK](https://cloud.google.com/sdk/) (125.0.0+)

## Labs

While GCP will be used for basic infrastructure needs, the things learned in this tutorial apply to every platform.

* [Cloud Infrastructure Provisioning](01-infrastructure.md)
* [Setting up a CA and TLS Cert Generation](02-certificate-authority.md)
* [Bootstrapping an H/A etcd cluster](03-etcd.md)
* [Bootstrapping an H/A Kubernetes Control Plane](04-kubernetes-controller.md)
* [Bootstrapping Kubernetes Workers](05-kubernetes-worker.md)
* [Configuring the Kubernetes Client - Remote Access](06-kubectl.md)
* [Managing the Container Network Routes](07-network.md)
* [Setting up cilium daemon-set](08-cilium.md)
* [Deploying the Cluster DNS Add-on](09-dns-addon.md)
* [Smoke Test](10-smoke-test.md)
* [Cleaning Up](11-cleanup.md)
