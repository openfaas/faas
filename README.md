## OpenFaaS &reg; - Serverless Functions Made Simple

[![Build Status](https://github.com/openfaas/faas/actions/workflows/build.yml/badge.svg)](https://github.com/openfaas/faas/actions/workflows/build.yml)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/openfaas/faas)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

![OpenFaaS Logo](https://blog.alexellis.io/content/images/2017/08/faas_side.png)

OpenFaaS&reg; makes it easy for developers to deploy event-driven functions and microservices to Kubernetes without repetitive, boiler-plate coding. Package your code or an existing binary in an OCI-compatible image to get a highly scalable endpoint with auto-scaling and metrics.

[![Twitter URL](https://img.shields.io/twitter/url/https/twitter.com/fold_left.svg?style=social&label=Follow%20%40openfaas)](https://twitter.com/openfaas)

**Highlights**

* Ease of use through UI portal and *one-click* install
* Write services and functions in any language with [Template Store](https://www.openfaas.com/blog/template-store/) or a Dockerfile
* Build and ship your code in an OCI-compatible/Docker image
* Portable: runs on existing hardware or public/private cloud by leveraging [Kubernetes](https://github.com/openfaas/faas-netes)
* [CLI](http://github.com/openfaas/faas-cli) available with YAML format for templating and defining functions
* Auto-scales as demand increases [including to zero](https://docs.openfaas.com/architecture/autoscaling/)
* [Commercially supported Pro distribution by the team behind OpenFaaS](https://openfaas.com/pricing/)

**Want to dig deeper into OpenFaaS?**

* Trigger endpoints with either [HTTP or events sources such as Apache Kafka and AWS SQS](https://docs.openfaas.com/reference/triggers/)
* Offload tasks to the built-in [queuing and background processing](https://docs.openfaas.com/reference/async/)
* Quick-start your Kubernetes journey with [GitOps from OpenFaaS Cloud](https://docs.openfaas.com/openfaas-cloud/intro/)
* Go secure or go home [with 5 must-know security tips](https://www.openfaas.com/blog/five-security-tips/)
* Learn everything you need to know to [go to production](https://docs.openfaas.com/architecture/production/)
* Integrate with Istio or Linkerd with [Featured Tutorials](https://docs.openfaas.com/tutorials/featured/#service-mesh)
* Deploy to [Kubernetes or OpenShift](https://docs.openfaas.com/deployment/)

## OpenFaaS Tiers and Pricing

This repository is part of OpenFaaS Community Edition (CE), which is licensed for non-commercial use by individuals, and a time-limited trial for commercial Proof Of Concepts (PoC). Internal use within a company or business requires a license.

OpenFaaS CE:

* has usage restrictions, which you can learn about in the [OpenFaaS CE EULA](EULA.md)
* has basic or primitive features and capabilities compared to the commercial versions
* is not licensed for commercial use of any kind beyond an initial trial period

OpenFaaS Standard and OpenFaaS for Enterprises are full and distinct commercial products. 

They are maintained and developed independently, by a full-time team, with commercial support, and active maintenance for CVEs, and updates in the Kubernetes and Cloud Native ecosystem.

Learn more about the tiers at [https://www.openfaas.com/pricing/](https://www.openfaas.com/pricing/)

## Overview of OpenFaaS (Serverless Functions Made Simple)

![Conceptual architecture](/docs/of-layer-overview.png)

> Conceptual architecture and stack, [more detail available in the docs](https://docs.openfaas.com/architecture/stack/)

### Code samples

You can scaffold a new function using the `faas-cli new` command passing in the name of the function and the language template you want to use i.e. `faas-cli new --lang node20 stripe-webhooks`.

Official templates exist for many popular languages and are easily extensible with Dockerfiles.

Learn about [OpenFaaS templates in the docs](https://docs.openfaas.com/languages/overview/)

* Node.js (`node20`) example:

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

* Python 3 example (`python3-http`):

    ```python
    def handle(event, context):
        return {
            "statusCode": 200,
            "body": "Hello from OpenFaaS!"
        }
    ```

    *handler.py*

* Golang example (`golang-middleware`)

    ```go
    package function
    
    import (
     	"fmt"
     	"io"
     	"net/http"
    )
    
    func Handle(w http.ResponseWriter, r *http.Request) {
   	    var input []byte
        
       	if r.Body != nil {
        		defer r.Body.Close()
        		body, _ := io.ReadAll(r.Body)
        		input = body
       	}
        
       	w.WriteHeader(http.StatusOK)
       	w.Write([]byte(fmt.Sprintf("Body: %s", string(input))))
    }
    ```

## Get started with OpenFaaS

### Official training resources

View our [official training materials](https://docs.openfaas.com/tutorials/training)

### Official eBook and video workshop
[![eBook logo](https://www.alexellis.io/serverless.png)](https://gumroad.com/l/serverless-for-everyone-else)

The founder of OpenFaaS wrote *Serverless For Everyone Else* to help developers understand the use-case for functions through practical hands-on exercises using JavaScript and Node.js. No programming experience is required to try the exercises.

The examples use the faasd project, which is an easy to use and lightweight way to start learning about OpenFaaS and functions.

[Check out Serverless For Everyone Else on Gumroad](https://gumroad.com/l/serverless-for-everyone-else)

### OpenFaaS and Golang

Everyday Go is a practical, hands-on guide to writing CLIs, web pages, and microservices in Go. It also features a chapter dedicated to development and testing of functions using OpenFaaS and Go.

* [Everyday Golang](https://openfaas.gumroad.com/l/everyday-golang)

### Community blog and documentation

* Read the documentation: [docs.openfaas.com](https://docs.openfaas.com/deployment)
* Read latest news and tutorials on the [Official Blog](https://www.openfaas.com/blog/)

### Quickstart

![OpenFaaS Community Edition UI](/docs/inception.png)

> Here is a screenshot of the OpenFaaS Community Edition UI which was designed for ease of use. The inception function is being run which is available on the in the store.

Deploy OpenFaaS to Kubernetes, OpenShift, or faasd now with a [deployment guide](https://docs.openfaas.com/deployment/)

### Video presentations

* [Meet faasd. Look Ma’ No Kubernetes! 2020](https://www.youtube.com/watch?v=ZnZJXI377ak&feature=youtu.be)
* [Getting Beyond FaaS: The PLONK Stack for Kubernetes Developers 2019](https://www.youtube.com/watch?v=NckMekZXRt8&feature=emb_title)
* [Serverless Beyond the Hype - Alex Ellis - GOTO 2018](https://www.youtube.com/watch?v=yOpYYYRuDQ0)
* [How LivePerson is Tailoring its Conversational Platform Using OpenFaaS - Simon Pelczer 2019](https://www.youtube.com/watch?v=bt06Z28uzPA)
* [Digital Transformation of Vision Banco Paraguay with Serverless Functions @ KubeCon 2018](https://kccna18.sched.com/event/GraO/digital-transformation-of-vision-banco-paraguay-with-serverless-functions-alex-ellis-vmware-patricio-diaz-vision-banco-saeca)
* [Introducing "faas" - Cool Hacks Keynote at Dockercon 2017](https://blog.docker.com/2017/04/dockercon-2017-mobys-cool-hack-sessions/)

### Community events and blog posts

Have you written a blog about OpenFaaS? Do you have a speaking event? Send a Pull Request to the community page below.

* [Read blogs/articles and find events about OpenFaaS](https://github.com/openfaas/faas/blob/master/community.md)

### Contributing

OpenFaaS Community Edition is written in Golang. All third-party contributions to the source code are made under the MIT license, additional restrictions apply to OpenFaaS CE as a whole, where contributions from OpenFaaS Ltd are licensed under the [OpenFaaS CE EULA](EULA.md). Various types of contributions are welcomed whether that means providing feedback, testing existing and new feature or hacking on the source code.

#### How do I become a contributor?

Please see the guide on [community & contributing](https://docs.openfaas.com/community/)

#### Dashboards

Example of a Grafana dashboard linked to OpenFaaS showing auto-scaling live in action: [here](https://grafana.com/dashboards/3526)

![OpenFaaS Pro auto-scaling dashboard with Grafana](https://pbs.twimg.com/media/FJ9EBVdWQAM9DeW?format=jpg&name=medium)
> [OpenFaaS Pro auto-scaling](https://docs.openfaas.com/architecture/autoscaling/) dashboard with Grafana

An alternative community dashboard is [available here](https://grafana.com/dashboards/3434)

### Press / Branding / Website Sponsorship

* Website Sponsorship 🌎

  If you'd like to gain visibility by displaying your logon on the [openfaas.com](https://www.openfaas.com/) homepage, feel free to reach out via email or browse the tiers via [GitHub Sponsors](https://github.com/sponsors/openfaas).

* Press / Analysts

  Looking at these repositories for commit counts and activity? All public repositories are part of OpenFaaS CE, a limited version of OpenFaaS aimed at giving people a low-barrier trial experience without having to sign up with a credit card. OpenFaaS CE is maintained on a best effort basis, but is not "OpenFaaS" itself. All OpenFaaS product development is done in private repositories, and cannot be tracked by third parties or by simply browsing GitHub.

  How are GitHub Stars and Forks counted? OpenFaaS CE is not a mono-repo, you cannot simply look at one repository and say "ah that's the count" - statistics are gathered from the whole [GitHub organisation](https://github.com/openfaas).

### Governance

OpenFaaS &reg; is an independent open-source project created by [Alex Ellis](https://www.alexellis.io), which is being built and shaped by a [growing community of contributors](https://www.openfaas.com/team/).

OpenFaaS is hosted by OpenFaaS Ltd (registration: 11076587), a company which also offers commercial services, homepage sponsorships, and support. OpenFaaS &reg; is a registered trademark in England and Wales.

### Users

View a selection of end-user companies who have given permission to have their logo listed at [openfaas.com](https://www.openfaas.com/).
