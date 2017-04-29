# Azure FaaS Cluster Template

This an experimental Azure template for deploying FaaS on live infra, based on a [barebones Swarm template][t] that deploys a master node and a dynamically scalable set of agent nodes. As configured, Azure will deploy more agents and expand infrastructure capacity as required by CPU load.

## Requirements

You need an Azure subscription, the `az` CLI, Python and the `make` command.

## TL;DR

    az login
    make keys
    make params
    make deploy-cluster
    make deploy-stack
    make proxy

    open http://localhost:8080

## How it Works

The cluster template defines a fixed set of master VMs (defaulting to 1) and an Azure VM scaleset (with 2 agent instances) that auto-scales based on CPU load, as well as a public load balancer mapped to the agent instances.

Both masters and agents are provisioned using `cloud-init`, which installs Docker CE and expands the Swarm cluster on the fly thanks to a set of simple helper scripts that make Swarm aware of when the infrastructure is scaling.

For security reasons, the admin and Prometheus UIs are only accessible via SSH tunnelling by default. Once the admin UI can be deployed separately, exposing functions via the public load balancer will be trivial.

[t]: https://github.com/rcarmo/azure-docker-swarm-cluster