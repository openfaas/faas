## OpenFaaS &reg; - Serverless Functions Made Simple

[![Go Report Card](https://goreportcard.com/badge/github.com/openfaas/faas)](https://goreportcard.com/report/github.com/openfaas/faas)
[![Build Status](https://travis-ci.com/openfaas/faas.svg?branch=master)](https://travis-ci.com/openfaas/faas)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/openfaas/faas)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)
[![Derek App](https://alexellis.o6s.io/badge?owner=openfaas&repo=faas)](https://github.com/alexellis/derek/)

![OpenFaaS Logo](https://blog.alexellis.io/content/images/2017/08/faas_side.png)

OpenFaaS&reg; makes it easy for developers to deploy event-driven functions and microservices to Kubernetes without repetitive, boiler-plate coding. Package your code or an existing binary in a Docker image to get a highly scalable endpoint with auto-scaling and metrics.

[![Twitter URL](https://img.shields.io/twitter/url/https/twitter.com/fold_left.svg?style=social&label=Follow%20%40openfaas)](https://twitter.com/openfaas)

**Highlights**

* Ease of use through UI portal and *one-click* install
* Write services and functions in any language with [Template Store](https://www.openfaas.com/blog/template-store/) or a Dockerfile
* Build and ship your code in the Docker/OCI image format
* Portable: runs on existing hardware or public/private cloud by leveraging [Kubernetes](https://github.com/openfaas/faas-netes)
* [CLI](http://github.com/openfaas/faas-cli) available with YAML format for templating and defining functions
* Auto-scales as demand increases [including to zero](https://www.openfaas.com/blog/zero-scale/)

**Want to dig deeper into OpenFaaS?**

* Trigger endpoints with either [HTTP or events sources such as AWS or Kafka](https://docs.openfaas.com/reference/triggers/)
* Offload tasks to the built-in [queuing and background processing](https://docs.openfaas.com/reference/async/)
* Quick-start your Kubernetes journey with [GitOps from OpenFaaS Cloud](https://docs.openfaas.com/openfaas-cloud/intro/)
* Go secure or go home [with 5 must-know security tips](https://www.openfaas.com/blog/five-security-tips/)
* Learn everything you need to know to [go to production](https://docs.openfaas.com/architecture/production/)
* Integrate with Istio or Linkerd with [Featured Tutorials](https://docs.openfaas.com/tutorials/featured/#service-mesh)
* Deploy to [Kubernetes or OpenShift](https://docs.openfaas.com/deployment/)

## Overview of OpenFaaS (Serverless Functions Made Simple)

![Conceptual architecture](/docs/of-layer-overview.png)

### Press / Branding / Website Sponsorship

* Individual Sponsorships / End-users / Insiders Track 🍻

  OpenFaaS is free to use and completely open source under the MIT license, however financial backing is required to sustain the effort to maintain and develop the project.
  
  Users and contributors are encouraged to join their peers in supporting the work through [GitHub Sponsors](https://insiders.openfaas.io/), all tiers gain exclusive access to the [Insiders Track](BACKERS.md)

* Website Sponsorship 🌎

  Companies and brands are welcome to sponsor [openfaas.com](https://www.openfaas.com/), the Gold and Platinum tiers come with a homepage logo, [see costs and tiers](BACKERS.md). Website sponsorships can be paid by invoice to OpenFaaS Ltd.

* Press / Branding 📸

  For information on branding, the press-kit, registered entities and sponsorship head over to the [openfaas/media](https://github.com/openfaas/media/blob/master/README.md) repo. You can also order custom SWAG or take part in the weekly Twitter contest [#FaaSFriday](https://twitter.com/search?q=faasfriday&src=typd)

  Looking for statistics? This project does not use a mono-repo, but is split across several components. Use [Ken Fukuyama's dashboard](https://kenfdev.o6s.io/github-stats-page) to gather accurate counts on contributors, stars and forks across the [GitHub organisation](https://github.com/openfaas).

  > Note: Incubator projects are not counted in these totals and are hosted under [openfaas-incubator](https://github.com/openfaas-incubator) awaiting graduation.

### Governance

OpenFaaS&reg; is an independent project founded by [Alex Ellis](https://www.alexellis.io) which is now being built and shaped by a growing community of contributors, GitHub Organisation members, Core contributors and end-users. OpenFaaS Ltd hosts the OpenFaaS codebase and trademarks. Professional services, sponsorships, and commercial support packages are available [upon request](mailto:sales@openfaas.com).

### Users

View a selection of end-user companies who have given permission to have their logo listed at [openfaas.com](https://www.openfaas.com/).

If you're using OpenFaaS please let us know [on this thread](https://github.com/openfaas/faas/issues/776). In addition, you are welcome to request to have your logo listed on the homepage. Thank you for your support.

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

Using the [faas-cli](http://github.com/openfaas/faas-cli) and `faas-cli up` you can create, build, distribute, and deploy your code in a very short period of time.

Any container or process in a Docker container can be a serverless workload in OpenFaaS, as long as it [conforms to The Workload Contract](https://docs.openfaas.com/reference/workloads/).

Create new functions from templates for Node.js, Python, [Go](https://blog.alexellis.io/serverless-golang-with-openfaas/) and many more. If you can't find a suitable template you can also use a Dockerfile or create your own.

The CLI is effectively a RESTful client for the API Gateway. When you have OpenFaaS configured you can [get started with the CLI here](https://blog.alexellis.io/quickstart-openfaas-cli/)

#### Examples

You can generate new functions using the `faas-cli` and built-in templates or use any binary for Windows or Linux in a Docker container.

Official templates exist for many popular languages and are easily extensible with Dockerfiles.

* Node.js (`node12`) example:

    ```js
   "use strict"
   
    module.exports = async (event, context) => {
        return context
            .status(200)
            .headers({"Content-Type": "text/html"})
            .succeed(`
            <h1>
                👋 Hello World 🌍
            </h1>`);
    }
 
    ```
    *handler.js*

* Python 3 example:

    ```python
    import requests

    def handle(req):
        r =  requests.get(req, timeout = 1)
        return "{} => {:d}".format(req, r.status_code)
    ```
    *handler.py*

* Golang example (`golang-http`)

    ```golang
    package function

    import (
        "log"

        "github.com/openfaas-incubator/go-function-sdk"
    )

    func Handle(req handler.Request) (handler.Response, error) {
        var err error

        return handler.Response{
            Body: []byte("Try us out today!"),
            Header: map[string][]string{
                "X-Served-By": []string{"openfaas.com"},
            },
        }, err
    }
    ```

The easiest way to get started with functions is to [take the workshop](https://github.com/openfaas/workshop) or one of the tutorials in the documentation.

## Get started with OpenFaaS

### Join the Community

* [Join Slack](https://docs.openfaas.com/community)
* [Become a GitHub sponsor](https://insiders.openfaas.io/)

### Official blog and documentation

* Read the documentation: [docs.openfaas.com](https://docs.openfaas.com/)
* Send a PR or raise an issue for the docs [openfaas/docs](https://github.com/openfaas/docs)
* Read latest news and tutorials on the [Official Blog](https://www.openfaas.com/blog/)

### Workshop (hands-on learning)

You can learn how to build functions and microservices with OpenFaaS using the [OpenFaaS workshop](http://github.com/openfaas/workshop). The Workshop has been curated carefully by dozens of community members to guide you through everything you need to know from setting up on Kubernetes to making use of secrets to integrating with GitHub and auto-scaling.

### Deploy OpenFaaS

Here is a screenshot of the API gateway portal - designed for ease of use with the inception function.

![Portal](/docs/inception.png)

#### Kubernetes

OpenFaaS is Kubernetes-native - you can follow the [deployment guide here](http://docs.openfaas.com/deployment/kubernetes/).

#### Docker Swarm

The deployment guide for Docker Swarm contains a simple one-line command to get you up and running in around 60 seconds. It also includes a set of [sample functions](https://github.com/openfaas/faas/tree/master/sample-functions) which you can use with the TestDrive instructions below.

[Deployment guide for Docker Swarm](http://docs.openfaas.com/deployment/docker-swarm/)

* Docker Playground

    You can quickly start OpenFaaS on Docker Swarm online using the community-run Docker playground: [Play-with-Docker](https://labs.play-with-docker.com/) (PWD)

    Simply follow the deployment guide for Swarm above in a new session

    > You will need a free Docker Hub account to get access. Get one here: [Docker Hub](https://hub.docker.com/)

## Video presentations

### OpenFaaS Cloud + Linkerd: A Secure, Multi-Tenant Serverless Platform

[From KubeCon North America 2019 with Charles Pretzer from Buoyant & Alex Ellis, OpenFaaS Ltd](https://www.youtube.com/watch?v=sD7hCwq3Gw0&feature=emb_title)

### The PLONK Stack/Serverless 2.0 for Kubernetes with OpenFaaS

[Getting Beyond FaaS: The PLONK Stack for Kubernetes Developers - Alex Ellis, OpenFaaS Ltd](https://www.youtube.com/watch?v=NckMekZXRt8&feature=emb_title)

### How LivePerson is Tailoring its Conversational Platform Using OpenFaaS @ KubeCon 2019

> Hear how LivePerson took one of the most popular open source Serverless projects (OpenFaaS) and built it into their product to add value for customers. Functions allow customers to create custom chatbot behaviour, messaging extensions and commerce workflows. You’ll see a live demo and hear about how the team put together the solution

* [Watch on YouTube](https://youtu.be/bt06Z28uzPA)

### Digital Transformation of Vision Banco Paraguay with Serverless Functions @ KubeCon late-2018

[From KubeCon North America 2018 with Alex and Patricio Diaz Senior Analyst, Vision Banco SAECA](https://kccna18.sched.com/event/GraO/digital-transformation-of-vision-banco-paraguay-with-serverless-functions-alex-ellis-vmware-patricio-diaz-vision-banco-saeca)

### Closing Keynote at Dockercon early-2017

Functions as a Service or FaaS was a winner in the Cool Hacks contest for Dockercon 2017.

* [Watch Alex present "FaaS" during the Dockercon 2017 keynote](https://blog.docker.com/2017/04/dockercon-2017-mobys-cool-hack-sessions/)

### Community events and blog posts

Have you written a blog about OpenFaaS? Do you have a speaking event? Send a Pull Request to the community page below.

* [Read blogs/articles and find events about OpenFaaS](https://github.com/openfaas/faas/blob/master/community.md)

If you'd like to join OpenFaaS community Slack channel to chat with contributors or get some help then check out [this page on community](https://docs.openfaas.com/community).

### Roadmap and contributing

OpenFaaS is written in Golang and is MIT licensed - contributions are welcomed whether that means providing feedback, testing existing and new feature or hacking on the source.

#### How do I become a contributor?

Please see the guide on [community & contributing](https://docs.openfaas.com/community/#contribute)

#### Roadmap

The roadmap for OpenFaaS is represented in [GitHub issues](https://github.com/openfaas/faas/issues) and [a Trello board](https://trello.com/b/5OpMyrBP/2020-openfaas-roadmap).

##### Roadmap: OpenFaaS Cloud

[OpenFaaS Cloud](https://github.com/openfaas/openfaas-cloud) is a platform built on top of the OpenFaaS framework which enables a multi-user experience driven by GitOps. It can be installed wherever you already have OpenFaaS and packages a dashboard along with CI/CD integration with GitHub so that you can push code to a private or public Git repo and get live HTTPS endpoints.

#### Dashboards

Example of a Grafana dashboards linked to OpenFaaS showing auto-scaling live in action: [here](https://grafana.com/dashboards/3526)

![](https://pbs.twimg.com/media/C9caE6CXUAAX_64.jpg:large)

An alternative community dashboard is [available here](https://grafana.com/dashboards/3434)
