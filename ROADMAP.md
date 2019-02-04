# Roadmap

## GitHub projects / source code 

You can find a detailed breakdown of the [openfaas](https://github.com/openfaas/) and [openfaas-incubator](https://github.com/openfaas-incubator/) organisations and projects [in the docs](https://docs.openfaas.com/contributing/get-started/).

## Feature overview

For an overview see [the docs](https://docs.openfaas.com/) or see a [feature comparison between OpenFaaS and OpenFaaS Cloud](https://docs.openfaas.com/openfaas-cloud/intro/).

### OpenFaaS

OpenFaaS is a platform for building Serverless Functions and/or deploying existing microservices. Any programming language or binary is supported with a range of [templates](https://github.com/openfaas/templates) available to help you get started.

The core services which make up OpenFaaS need to run on a Linux master, but Windows worker nodes can be added to your cluster to run Windows binaries and functions.

Platforms: the x86_64 platform has first class support, with 32-bit arm and 64-bit arm provided on a best-effort basis.

Orchestrators: there is official support for Kubernetes & Docker Swarm with the community providing support for AWS Fargate, Hashicorp Nomad and others.

### OpenFaaS Cloud

OpenFaaS Cloud is a multi-user distribution of OpenFaaS with a built-in CI/CD pipeline, OAuth delegation, a dashboard and a git-based workflow with public/private GitHub and self-hosted GitLab.

Platforms: only the x86_64 platform has support.

Orchestrators: there is official support for Kubernetes and for Docker Swarm. Kubernetes has better support for multiple users / tenants than Swarm.

## What is coming next?

Proposals and feature requests are tracked in Trello and through the GitHub issue tracker of each project in the two organisations.

* [openfaas](https://github.com/openfaas/)
* [openfaas-incubator](https://github.com/openfaas-incubator/)

## Contributing

Please see [CONTRIBUTING.md](https://github.com/openfaas/faas/blob/master/CONTRIBUTING.md).
