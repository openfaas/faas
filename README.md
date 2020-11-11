## OpenFaaS &reg; - Serverless Functions Made Simple

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

> Conceptual architecture and stack, [more detail available in the docs](https://docs.openfaas.com/architecture/stack/)

### Press / Branding / Website Sponsorship

* Individual Sponsorships / End-users / Insiders Track 🍻

    The source code for OpenFaaS is free to use and open source under the terms of the MIT license.

    OpenFaaS Ltd offers [commercial support and enterprise add-ons](https://www.openfaas.com/support) for end-users and [training and consulting services for Cloud and Kubernetes](https://www.openfaas.com/consulting).

    Users and contributors are encouraged to join their peers in supporting the project through [GitHub Sponsors](https://www.openfaas.com/support).

* Website Sponsorship 🌎

  Companies and brands are welcome to sponsor [openfaas.com](https://www.openfaas.com/), the Gold and Platinum tiers come with a homepage logo, [see costs and tiers](BACKERS.md). Website sponsorships are payable by invoice.

* Press / Branding 📸

  For information on branding, the press-kit, registered entities and sponsorship head over to the [openfaas/media](https://github.com/openfaas/media/blob/master/README.md) repo. You can also order custom SWAG or take part in the weekly Twitter contest [#FaaSFriday](https://twitter.com/search?q=faasfriday&src=typd)

  Looking for statistics? This project does not use a mono-repo, but is split across several components. Use [Ken Fukuyama's dashboard](https://kenfdev.o6s.io/github-stats-page) to gather accurate counts on contributors, stars and forks across the [GitHub organisation](https://github.com/openfaas).

  > Note: any statistics you gather about the openfaas/faas repository will be invalid, the faas repo is not representative of the project's activity.

### Governance

OpenFaaS &reg; is an independent open-source project created by [Alex Ellis](https://www.alexellis.io), which is being built and shaped by a [growing community of contributors](https://www.openfaas.com/team/).

OpenFaaS is hosted by OpenFaaS Ltd (registration: 11076587), a company which also offers commercial services, homepage sponsorships, and support. OpenFaaS &reg; is a registered trademark in England and Wales.

### Users

View a selection of end-user companies who have given permission to have their logo listed at [openfaas.com](https://www.openfaas.com/).

If you're using OpenFaaS please let us know [on this thread](https://github.com/openfaas/faas/issues/776). In addition, you are welcome to request to have your logo listed on the homepage. Thank you for your support.

### Code samples

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

## Get started with OpenFaaS

### Official blog and documentation

* Read the documentation: [docs.openfaas.com](https://docs.openfaas.com/deployment)
* Read latest news and tutorials on the [Official Blog](https://www.openfaas.com/blog/)

## Community Subscription

OpenFaaS users can subscribe to a weekly Community Newsletter called Insiders Updates, to keep up to date with new features, bug fixes, events, tutorials and security patches. Insiders Updates are written by the project founder and distributed via GitHub Sponsors.

* [Get a Community Subscription](https://github.com/support/)

### Support & getting help

* [Cloud Native Consulting](https://www.openfaas.com/consulting) - get hands-on expert help with your cloud, Kubernetes and OpenFaaS migration and projects
* [Commercial support](https://www.openfaas.com/support) - a subscription service from OpenFaaS Ltd
* [Join Slack](https://docs.openfaas.com/community) - run by community volunteers

### Online training

* **New**: Training course from the LinuxFoundation: Introduction to Serverless on Kubernetes

    This training course "Introduction to Serverless on Kubernetes" written by the project founder and commissioned by the LinuxFoundation provides an overview of what you need to know to build functions and operate OpenFaaS on public cloud.

    Training course: [Introduction to Serverless on Kubernetes](https://www.edx.org/course/introduction-to-serverless-on-kubernetes)

* Self-paced workshop written by the community on GitHub

    You may also like to try the self-paced workshop on GitHub written by the OpenFaaS community

    Browse the [workshop](https://github.com/openfaas/workshop)

* Corporate trainings

    If you wish to arrange a training session for your team, or a consultation, [feel free to contact OpenFaaS Ltd](https://www.openfaas.com/support/)

### Quickstart


![Portal](/docs/inception.png)

> Here is a screenshot of the API gateway portal - designed for ease of use with the inception function.

Deploy OpenFaaS to Kubernetes, OpenShift, or faasd [deployment guides](./deployment/)

### Video presentations

* [Meet faasd. Look Ma’ No Kubernetes!](https://www.youtube.com/watch?v=ZnZJXI377ak&feature=youtu.be)
* [OpenFaaS Cloud + Linkerd: A Secure, Multi-Tenant Serverless Platform](https://www.youtube.com/watch?v=sD7hCwq3Gw0&feature=emb_title)
* [Getting Beyond FaaS: The PLONK Stack for Kubernetes Developers](https://www.youtube.com/watch?v=NckMekZXRt8&feature=emb_title)
* [Digital Transformation of Vision Banco Paraguay with Serverless Functions @ KubeCon late-2018](https://kccna18.sched.com/event/GraO/digital-transformation-of-vision-banco-paraguay-with-serverless-functions-alex-ellis-vmware-patricio-diaz-vision-banco-saeca)
* [Introducing "faas" - Cool Hacks Keynote at Dockercon 2017](https://blog.docker.com/2017/04/dockercon-2017-mobys-cool-hack-sessions/)

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
