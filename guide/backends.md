## OpenFaaS backends

OpenFaaS is a framework for building serverless functions with containers and running them at scale.

We support two orchestration platforms or "backends":

* Docker Swarm
* Kubernetes

The Docker Swarm support is built-into the faas repo, but the Kubernetes support is provided by a microservice in the [faas-netes](https://github.com/alexellis/faas-netes) repo.

### I need a backend for X

This project is focusing on Docker Swarm and Kubernetes, but we're open to support from third parties and vendors for other backends:

Here are some ideas:

* Nomad
* Marathon Mesos
* AWS ECS
* Hyper.sh

If you would like to write your own back-end for `X` then you can write your own microservice that conforms to the [Swagger API](https://github.com/alexellis/faas/tree/master/api-docs) here.

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

* See the [FaaS-netes handlers](https://github.com/alexellis/faas-netes/tree/master/handlers) for examples of how to implement each endpoint.

* See the [Swagger API](https://github.com/alexellis/faas/tree/master/api-docs)

