## OpenFaaS backends guide

OpenFaaS is a framework for building serverless functions with containers and running them at scale.

> Bring Serverless OpenFaaS functions to your favourite container platform.

We support two orchestration platforms or "backends":

* Docker Swarm
* Kubernetes

There is also community work in-progress to support:

* Rancher/Cattle

The Docker Swarm support is built-into the faas repo, but the Kubernetes support is provided by a microservice in the [faas-netes](https://github.com/alexellis/faas-netes) repo.

If you're thinking of writing a new back-end we'd love to hear about it and help you, so please get in touch with alex@openfaas.com. Existing implementations (like OpenFaaS) are written in Golang and this provides a level of consistency across the projects.

### I need a backend for X

This project is focusing on Docker Swarm and Kubernetes, but we're open to support from third parties and vendors for other backends:

Here are some ideas:

* Nomad
* Marathon Mesos
* AWS ECS
* Hyper.sh

If you would like to write your own back-end for `X` then you can write your own microservice that conforms to the [Swagger API](https://github.com/alexellis/faas/tree/master/api-docs) here.

### How does my back-end work?

In order to support a new back end you will write a new "external_provider" and configure this on the API Gateway. The API Gateway will then proxy any requests to your new microservice. The first "external_provider" was the Kubernetes implementation [faas-netes](https://github.com/alexellis/faas-netes):

![](https://camo.githubusercontent.com/c250e0dc975e50b6fae3fc84ba5cdc0274bc305c/68747470733a2f2f7062732e7477696d672e636f6d2f6d656469612f44466837692d5a586b41415a6b77342e6a70673a6c61726765)

Deploy a function - through the built-in Swarm support or through faas-netes

![](https://pbs.twimg.com/media/DIyFFnsXkAAa5Gj.jpg)

Invoke your function - through the built-in Swarm or via faas-netes

![](https://pbs.twimg.com/media/DIyFFnqXgAAMyCh.jpg)

Find out more about the [watchdog here](https://github.com/alexellis/faas/tree/master/watchdog).

### Automatically compatible OpenFaaS

The following are fully compatible with any additional back-ends:

* API Gateway
* Promethes metrics (tracked through API Gateway)
* The built-in UI portal (hosted on the API Gateway)
* The Function Watchdog and any existing OpenFaaS functions
* The [CLI](https://github.com/alexellis/faas-cli)
* Asynchronous function invocation

Dependent on back-end:

* Secrets or environmental variable support
* Windows Containers function runtimes (i.e. via W2016 and Docker)
* Scaling - dependent on underlying API (available in Docker & Kubernetes)

#### Backend endpoints:

* List / Create / Delete a function

`/system/functions`

Method(s): GET / POST / DELETE 

* Get a specific function

`/system/function/{name:[-a-zA-Z_0-9]+}`

Method(s): GET

* Scale a specific function:

`/system/scale-function/{name:[-a-zA-Z_0-9]+}`

Method(s): POST

* Invoke a specific function

`/function/{name:[-a-zA-Z_0-9]+}`

Method(s): POST


### Examples / documentation

* See the [Swagger API](https://github.com/alexellis/faas/tree/master/api-docs) as a starting point.

#### faas-netes (Kubernetes)

The Kubernetes integration was written by Alex Ellis and is officially supported by the project.

* See the [FaaS-netes handlers](https://github.com/alexellis/faas-netes/tree/master/handlers) for examples of how to implement each endpoint.

#### Rancher / Cattle (community)

This work is by Ken Fukuyama from Japan.

* [Blog post](https://medium.com/@kenfdev/openfaas-on-rancher-684650cc078e)

* [faas-rancher](https://github.com/kenfdev/faas-rancher) implementation in Golang
