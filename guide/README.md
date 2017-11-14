OpenFaaS guides
================

This page is a collection of our key blog posts, tutorials and guides while we prepare a [dedicated site](https://github.com/openfaas/faas/issues/253) for documentation. For other queries please get in touch for a Slack invite or ping [@openfaas](https://twitter.com/openfaas) on Twitter.

> There is a PR underway for the new [documentation site](https://github.com/openfaas/faas/pull/274)

Suggestions for new topics are welcome. Please also check the [Issue tracker](https://github.com/openfaas/faas/issues).

## Deployment guides (start here)

### A foreword on security

These instructions are for a development environment. If you plan to expose OpenFaaS on the public Internet you need to enable basic authentication with a proxy such as Kong or Traefik at a minimum. TLS is also highly recomended and freely available with LetsEncrypt.org. [Kong guide](https://github.com/openfaas/faas/blob/master/guide/kong_integration.md) [Traefik guide](https://github.com/openfaas/faas/blob/master/guide/traefik_integration.md).

> Note: We are also looking to [automate authentication "out the box"](https://github.com/openfaas/faas/issues/349) to cover edge cases.

* [Kubernetes deployment](deployment_k8s.md)

* [Docker Swarm deployment](deployment_swarm.md)

* [DigitalOcean deployment (with Swarm)](deployment_digitalocean.md)

## Intermediate

* [Workflows / Chaining functions](chaining_functions.md)

* [Interacting with other containers/services](interactions.md)

* [Troubleshooting](troubleshooting.md)

* [Asynchronous functions with NATS Streaming](asynchronous.md)

* [Hardening OpenFaaS with Kong & TLS](kong_integration.md)

* [Reference documentation for Function Watchdog](../watchdog/)

* WIP [Debugging Functions](https://github.com/openfaas/faas/issues/223)

## Blog posts and tutorials

### Hands-on with Node.js / Go / Python

* [Morning coffee with the OpenFaaS CLI](https://blog.alexellis.io/quickstart-openfaas-cli/)

* [Build a Serverless Golang Function with OpenFaaS](https://blog.alexellis.io/serverless-golang-with-openfaas/)

* [Your first serverless Python function with OpenFaaS](https://blog.alexellis.io/first-faas-python-function/)

### Project background, design decisions and architecture:

* [Introducing Functions as a Service (OpenFaaS)](https://blog.alexellis.io/introducing-functions-as-a-service/)

* [OpenFaaS presents to CNCF Serverless workgroup](https://blog.alexellis.io/openfaas-cncf-workgroup/)

* [An Introduction to Serverless DevOps with OpenFaaS](https://hackernoon.com/an-introduction-to-serverless-devops-with-openfaas-b978ab0eb2b)

### Hands-on with containers as functions

* [Serverless sorcery with ImageMagick](https://blog.alexellis.io/serverless-imagemagick/)

### Fine-tuning / high-throughput

* [OpenFaaS accelerates serverless Java with AfterBurn](https://blog.alexellis.io/openfaas-serverless-acceleration/)

### Raspberry Pi & ARM

[Your Serverless Raspberry Pi cluster with Docker](https://blog.alexellis.io/your-serverless-raspberry-pi-cluster/)

## Extend OpenFaaS

* [Build a third-party provider](backends.md)
