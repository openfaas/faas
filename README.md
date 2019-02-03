## OpenFaaS - Serverless Functions Made Simple

[![Go Report Card](https://goreportcard.com/badge/github.com/openfaas/faas)](https://goreportcard.com/report/github.com/openfaas/faas) [![Build
Status](https://travis-ci.org/openfaas/faas.svg?branch=master)](https://travis-ci.org/openfaas/faas) [![GoDoc](https://godoc.org/github.com/openfaas/faas?status.svg)](https://godoc.org/github.com/openfaas/faas) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

![OpenFaaS Logo](https://blog.alexellis.io/content/images/2017/08/faas_side.png)

OpenFaaS&reg; (Functions as a Service) is a framework for building Serverless functions with Docker and Kubernetes which has first-class support for metrics. Any process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

[![Twitter URL](https://img.shields.io/twitter/url/https/twitter.com/fold_left.svg?style=social&label=Follow%20%40openfaas)](https://twitter.com/openfaas)

**Highlights**

* Ease of use through UI portal and *one-click* install
* Write functions in any language for Linux or Windows and package in Docker/OCI image format
* Portable - runs on existing hardware or public/private cloud - [Kubernetes](https://github.com/openfaas/faas-netes) and Docker Swarm native
* [CLI](http://github.com/openfaas/faas-cli) available with YAML format for templating and defining functions
* Auto-scales as demand increases [including to zero](https://www.openfaas.com/blog/zero-scale/)

## Overview of OpenFaaS

> Serverless Functions Made Simple.

![Stack](https://pbs.twimg.com/media/DFrkF4NXoAAJwN2.jpg)

### Press / Branding / Sponsors

* Press / Branding

  For information on branding, the press-kit, registered entities and sponsorship head over to the [openfaas/media](https://github.com/openfaas/media/blob/master/README.md) repo. You can also order custom SWAG or take part in the weekly Twitter contest [#FaaSFriday](https://twitter.com/search?q=faasfriday&src=typd)

* Looking for statistics? This project does not use a mono-repo, but is split across several components. Use [Ken Fukuyama's dashboard](https://kenfdev.o6s.io/github-stats-page) to gather accurate counts on contributors, stars and forks across the [GitHub organisation](https://github.com/openfaas).

  > Note: Incubator projects are not counted in these totals and are hosted under [openfaas-incubator](https://github.com/openfaas-incubator) awaiting graduation.

* Support for OpenFaaS
  OpenFaaS is free to use and completely open source under the MIT license. You can donate to the project to fund its ongoing development or become a sponsor. [Support OpenFaaS](https://www.openfaas.com/donate/)

### Governance

OpenFaaS&reg; is an independent project founded by [Alex Ellis](https://www.alexellis.io) which is now being built and shaped by a growing community of contributors, GitHub Organisation members, Core contributors and end-users. More at: [openfaas.com](https://www.openfaas.com).

### Users

[View our end-users](https://docs.openfaas.com/#users-of-openfaas) or get in touch to [have your company added](https://github.com/openfaas/faas/issues/776).

> Please support [OpenFaaS on Patreon](https://www.patreon.com/alexellis)) and back a great community at the same time. You will be listed as a [backers or sponsor here](https://github.com/openfaas/faas/blob/master/BACKERS.md).

Thank you for your support.

### Technical overview

#### Function Watchdog

* You can make any Docker image into a serverless function by adding the *Function Watchdog* (a tiny Golang HTTP server)
* The *Function Watchdog* is the entrypoint allowing HTTP requests to be forwarded to the target process via STDIN or HTTP. The response is sent back to the caller by writing to STDOUT or HTTP from your application.

#### API Gateway / UI Portal

* The API Gateway provides an external route into your functions and collects Cloud Native metrics through Prometheus.
* Your API Gateway will scale functions according to demand by altering the service replica count in the Docker Swarm or Kubernetes API.
* A UI is baked in allowing you to invoke functions in your browser and create new ones as needed.

> The API Gateway is a RESTful micro-service and you can view the [Swagger docs here](https://github.com/openfaas/faas/tree/master/api-docs).

#### CLI

Any container or process in a Docker container can be a serverless function in FaaS. Using the [FaaS CLI](http://github.com/openfaas/faas-cli) you can deploy your functions quickly.

Create new functions from templates for Node.js, Python, [Go](https://blog.alexellis.io/serverless-golang-with-openfaas/) and many more. If you can't find a suitable template you can also use a Dockerfile.

> The CLI is effectively a RESTful client for the API Gateway.

When you have OpenFaaS configured you can [get started with the CLI here](https://blog.alexellis.io/quickstart-openfaas-cli/)

#### Function examples

You can generate new functions using the FaaS-CLI and built-in templates or use any binary for Windows or Linux in a Docker container.

Official templates exist for many popular languages and are easily extensible with Dockerfiles. Here is an example with Python 3 and Node.js:

* Python 3 example:

```python
import requests

def handle(req):
    r =  requests.get(req, timeout = 1)
    return "{} => {:d}".format(req, r.status_code)
```
*handler.py*

* Node.js example:

```js
"use strict"

module.exports = (callback, context) => {
    var err;
    callback(err, {"message": "You said: " + context})
}
```
*handler.js*

The easiest way to get started with functions is to take the workshop or one of the tutorials in the documentation.

## Get started with OpenFaaS

### Official documentation and blog

See our documentation on [docs.openfaas.com](https://docs.openfaas.com/). The source repository for the documentation website is [openfaas/docs](https://github.com/openfaas/docs).

Read latest news on OpenFaaS from the community [blog](https://www.openfaas.com/blog/)

### Hands-on labs (detailed getting started)

You can learn how to build functions with OpenFaaS using our hands-on labs in the [OpenFaaS workshop](http://github.com/openfaas/workshop).

### TestDrive (classic getting started)

**Kubernetes**

OpenFaaS is Kubernetes-native - you can follow the [deployment guide here](http://docs.openfaas.com/deployment/kubernetes/).

**Docker Swarm**

The deployment guide for Docker Swarm contains a simple one-line command to get you up and running in around 60 seconds. It also includes a set of [sample functions](https://github.com/openfaas/faas/tree/master/sample-functions) which you can use with the TestDrive instructions below.

[Deployment guide for Docker Swarm](http://docs.openfaas.com/deployment/docker-swarm/)

**Docker Playground**

You can quickly start OpenFaaS on Docker Swarm online using the community-run Docker playground: [Play-with-Docker](https://labs.play-with-docker.com/) (PWD)

Simply follow the deployment guide for Swarm above in a new session

> You will need a free Docker Hub account to get access. Get one here: [Docker Hub](https://hub.docker.com/)

#### Begin the TestDrive

* [Begin the TestDrive with Docker Swarm](https://github.com/openfaas/faas/blob/master/TestDrive.md)

Here is a screenshot of the API gateway portal - designed for ease of use.

![Portal](https://pbs.twimg.com/media/C7bkpZbWwAAnKsx.jpg)

## Find out more about OpenFaaS

### Digital Transformation of Vision Banco Paraguay with Serverless Functions @ KubeCon late-2018

[HD video co-presenting at KubeCon with Patricio Diaz Senior Analyst, Vision Banco SAECA](https://kccna18.sched.com/event/GraO/digital-transformation-of-vision-banco-paraguay-with-serverless-functions-alex-ellis-vmware-patricio-diaz-vision-banco-saeca)

### Serverless Beyond the Hype (goto Copenhagen) late-2018

Overview of the Serverless landscape for Kubernetes, OpenFaaS and OpenFaaS Cloud with live demos and most update information. [View on Android or iPhone](https://gotocph.com/2018/sessions/592)

### The Cube interview @ DevNet Create mid-2018

* [2018 update on the OpenFaaS story](https://www.youtube.com/watch?v=J8UYZ1GXNTQ)

### TechFieldDay presentation (Dockercon EU) late-2017

15 minute overview with demos on Kubernetes and with Alexa - [HD YouTube video](https://www.youtube.com/watch?v=C3agSKv2s_w&list=PLlIapFDp305AiwA17mUNtgi5-u23eHm5j&index=1)

### Closing Keynote at Dockercon early-2017

Functions as a Service or FaaS was a winner in the Cool Hacks contest for Dockercon 2017.

* [Watch my FaaS keynote at Dockercon 2017](https://blog.docker.com/2017/04/dockercon-2017-mobys-cool-hack-sessions/)

If you'd like to find the functions I used in the demos head over to the [faas-dockercon](https://github.com/alexellis/faas-dockercon/) repository.

**Background story**

* [Introducing OpenFaaS (Functions as a Service)](https://blog.alexellis.io/introducing-functions-as-a-service/) -  August 2017
* [Functions as a Service (FaaS)](http://blog.alexellis.io/functions-as-a-service/) - January 2017

### Community

Have you written a blog about OpenFaaS? Send a Pull Request to the community page below.

* [Read blogs/articles and find events about OpenFaaS](https://github.com/openfaas/faas/blob/master/community.md)

If you'd like to join OpenFaaS community Slack channel to chat with contributors or get some help then check out [this page on community](https://docs.openfaas.com/community).

### Roadmap and contributing

OpenFaaS is written in Golang and is MIT licensed - contributions are welcomed whether that means providing feedback, testing existing and new feature or hacking on the source.

#### How do I become a contributor?

Please see the guide on [community & contributing](https://docs.openfaas.com/community/#contribute)

#### Roadmap

The roadmap for OpenFaaS is represented in [GitHub issues](https://github.com/openfaas/faas/issues) and a Trello board. There is also a historical ROADMAP file in the [main faas repository](https://github.com/openfaas/faas/blob/master/ROADMAP.md).

##### Roadmap: OpenFaaS Cloud

[OpenFaaS Cloud](https://github.com/openfaas/openfaas-cloud) is a platform built on top of the OpenFaaS framework which enables a multi-user experience driven by GitOps. It can be installed wherever you already have OpenFaaS and packages a dashboard along with CI/CD integration with GitHub so that you can push code to a private or public Git repo and get live HTTPS endpoints.

#### Dashboards

Example of a Grafana dashboards linked to OpenFaaS showing auto-scaling live in action: [here](https://grafana.com/dashboards/3526)

![](https://pbs.twimg.com/media/C9caE6CXUAAX_64.jpg:large)

An alternative community dashboard is [available here](https://grafana.com/dashboards/3434)
