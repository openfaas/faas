## OpenFaaS - Functions as a Service

[![Go Report Card](https://goreportcard.com/badge/github.com/openfaas/faas)](https://goreportcard.com/report/github.com/openfaas/faas) [![Build
Status](https://travis-ci.org/openfaas/faas.svg?branch=master)](https://travis-ci.org/openfaas/faas) [![GoDoc](https://godoc.org/github.com/openfaas/faas?status.svg)](https://godoc.org/github.com/openfaas/faas)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

![OpenFaaS Logo](https://blog.alexellis.io/content/images/2017/08/faas_side.png)

OpenFaaS (Functions as a Service) is a framework for building serverless functions with Docker which has first class support for metrics. Any process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

[![Twitter URL](https://img.shields.io/twitter/url/https/twitter.com/fold_left.svg?style=social&label=Follow%20%40open_faas)](https://twitter.com/open_faas)

**Highlights**

* Ease of use through UI portal and *one-click* install
* Write functions in any language for Linux or Windows and package in Docker/OCI image format
* Portable - runs on existing hardware or public/private cloud - [Kubernetes](https://github.com/openfaas/faas-netes) and Docker Swarm native
* [CLI](http://github.com/openfaas/faas-cli) available with YAML format for templating and defining functions
* Auto-scales as demand increases

## Governance

OpenFaaS is an independent project created by [Alex Ellis](https://www.alexellis.io) which is now being built and shaped by a growing community of contributors. Project website: [openfaas.com](https://www.openfaas.com).

## Overview of OpenFaaS

![Stack](https://pbs.twimg.com/media/DFrkF4NXoAAJwN2.jpg)

### Function Watchdog

* You can make any Docker image into a serverless function by adding the *Function Watchdog* (a tiny Golang HTTP server)
* The *Function Watchdog* is the entrypoint allowing HTTP requests to be forwarded to the target process via STDIN. The response is sent back to the caller by writing to STDOUT from your application.

### API Gateway / UI Portal

* The API Gateway provides an external route into your functions and collects Cloud Native metrics through Prometheus.
* Your API Gateway will scale functions according to demand by altering the service replica count in the Docker Swarm or Kubernetes API.
* A UI is baked in allowing you to invoke functions in your browser and create new ones as needed.

> The API Gateway is a RESTful micro-service and you can view the [Swagger docs here](https://github.com/openfaas/faas/tree/master/api-docs).

### CLI

Any container or process in a Docker container can be a serverless function in FaaS. Using the [FaaS CLI](http://github.com/openfaas/faas-cli) you can deploy your functions or quickly create new functions from templates such as Node.js or Python.

> The CLI is effectively a RESTful client for the API Gateway.

When you have OpenFaaS configured you can [get started with the CLI here](https://blog.alexellis.io/quickstart-openfaas-cli/)

### Function examples

You can generate new functions using the FaaS-CLI and built-in templates or use any binary for Windows or Linux in a Docker container.

* Python example:

```python
import requests

def handle(req):
        r =  requests.get(req, timeout = 1)
        print(req +" => " + str(r.status_code))
```
*handler.py*

* Node.js example:

```js
"use strict"

module.exports = (callback, context) => {
    callback(null, {"message": "You said: " + context})
}
```
*handler.js*

Other [Sample functions](https://github.com/openfaas/faas/tree/master/sample-functions) are available in the Github repository in a range of programming languages.

## Get started with OpenFaaS

### TestDrive

**Docker Swarm**

The deployment guide for Docker Swarm contains a simple one-line command to get you up and running in around 60 seconds. It also includes a set of [sample functions](https://github.com/openfaas/faas/tree/master/sample-functions) which you can use with the TestDrive instructions below.

[Deployment guide for Docker Swarm](https://github.com/openfaas/faas/blob/master/guide/deployment_swarm.md)

**Kubernetes**

OpenFaaS is Kubernetes-native - you can follow the [deployment guide here](https://github.com/openfaas/faas/blob/master/guide/deployment_k8s.md).

**Docker Playground**

You can quickly start OpenFaaS on Docker Swarm online using the community-run Docker playground: play-with-docker.com (PWD) by clicking the button below:

[![Try in PWD](https://cdn.rawgit.com/play-with-docker/stacks/cff22438/assets/images/button.png)](http://play-with-docker.com?stack=https://raw.githubusercontent.com/openfaas/faas/master/docker-compose.yml&stack_name=func)

#### Begin the TestDrive

* [Begin the TestDrive with Docker Swarm](https://github.com/openfaas/faas/blob/master/TestDrive.md)

Here is a screenshot of the API gateway portal - designed for ease of use.

![Portal](https://pbs.twimg.com/media/C7bkpZbWwAAnKsx.jpg)

## Find out more about OpenFaaS

### SkillsMatter video presentation

Great overview of OpenFaaS features, users and roadmap

* [HD Video](https://skillsmatter.com/skillscasts/10813-faas-and-furious-0-to-serverless-in-60-seconds-anywhere)

### OpenFaaS presents to CNCF Serverless workgroup

* [Video and blog post](https://blog.alexellis.io/openfaas-cncf-workgroup/)

### Closing Keynote at Dockercon 2017

Functions as a Service or FaaS was a winner in the Cool Hacks contest for Dockercon 2017.

* [Watch my FaaS keynote at Dockercon 2017](https://blog.docker.com/2017/04/dockercon-2017-mobys-cool-hack-sessions/)

If you'd like to find the functions I used in the demos head over to the [faas-dockercon](https://github.com/alexellis/faas-dockercon/) repository.

**Background story**

This is my original blog post on FaaS from January: [Functions as a Service blog post](http://blog.alexellis.io/functions-as-a-service/)

### Community

Have you written a blog about OpenFaaS? Send a Pull Request to the community page below.

* [Read blogs/articles and find events about OpenFaaS](https://github.com/openfaas/faas/blob/master/community.md)

If you'd like to join OpenFaaS community Slack channel to chat with contributors or get some help - then send a Tweet to [@alexellisuk](https://twitter.com/alexellisuk/) or email alex@openfaas.com.

### Roadmap and contributing

OpenFaaS is written in Golang and is MIT licensed - contributions are welcomed whether that means providing feedback, testing existing and new feature or hacking on the source.

To get started you can read the [roadmap](https://github.com/openfaas/faas/blob/master/ROADMAP.md) and [contribution guide](https://github.com/openfaas/faas/blob/master/CONTRIBUTING.md) or:

* [Browse FaaS issues on Github](https://github.com/openfaas/faas/issues).
* [Browse FaaS-CLI issues on Github](https://github.com/openfaas/faas-cli/issues).

Highlights:

* New: Kubernetes support via [FaaS-netes](https://github.com/openfaas/faas-netes) plugin
* New: FaaS CLI and easy install via `curl` and `brew`
* New: Windows function support
* New: Asynchronous/long-running OpenFaaS functions via [NATS Streaming](https://nats.io/documentation/streaming/nats-streaming-intro/) - [Follow this guide](https://github.com/openfaas/faas/blob/master/guide/asynchronous.md)

### How do I become a contributor?

Anyone is invited to contribute to the project in-line with the [contribution guide](https://github.com/openfaas/faas/blob/master/CONTRIBUTING.md) - you can also read the guide for ideas on how to get involved. We invite new contributors to join our Slack community. We would also ask you to propose any changes or contributions ahead of time, especially when there is no issue or proposal alredy tracking it.

### Other

Example of a Grafana dashboard linked to OpenFaaS showing auto-scaling live in action:

![](https://pbs.twimg.com/media/C9caE6CXUAAX_64.jpg:large)

Sample dashboard JSON file available [here](https://github.com/openfaas/faas/blob/master/contrib/grafana.json)
