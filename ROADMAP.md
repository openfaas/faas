# Roadmap

## 1. Current features

For an overview of features in August 2017 read the following post:

* [Introducing Functions as a Service (FaaS)](https://blog.alexellis.io/introducing-functions-as-a-service/)

## GitHub repos:

* https://github.com/alexellis/faas
* https://github.com/alexellis/faas-netes
* https://github.com/alexellis/faas-cli
* https://github.com/openfaas/nats-queue-worker

### The API Gateway

This container acts in a similar way to the API Gateway on AWS. Requests can be made to this endpoint with a JSON body.

Some of the more recent Completed Features:

* UI for viewing and testing functions deployed through stack
* Auto-scaling of replicas as load increases
* Backing off of replicas as load reduces
* Unique URL routes for serverless functions
* Instrumentation via Prometheus metrics at GET /metrics
* Bundled Prometheus stack with AlertManager
* UI enhancements to create new function through a form
* Raspberry Pi (armhf/armv6) support (currently available)
* Documentation for current API in Swagger format
* Documentation about [Alexa sample function](https://blog.alexellis.io/serverless-alexa-skill-mobymingle/)
* Native CLI for templating/building and deploying functions
* Windows support for watchdog back-end - so that Windows executables can be used in a multi-OS swarm
* Enforcing function execution time in seconds.
* Python, Node.js, Ruby and CSharp code templates for the CLI
* Delete function in CLI
* Developer guide for CSharp
* Developer guide for Python
* **Kubernetes support**
* Asynchronous / long-running tasks via NATS Streaming
* CLI - invoke / list functions / query function info
* OS constraints in the deploy function API

**Incoming requests and routing**

There are three options for routing:

* Functions created on the overlay network can be invoked by: http://localhost:8080/function/{servicename}
* Automatic routing is also enabled through the `/` endpoint via a `X-Function` header which matches a service name (function) directly.

### The watchdog

This binary fwatchdog acts as a watchdog for your function. Features:

* Static binary in Go
* Listens to HTTP requests over swarm overlay network
* Spawns process set in `fprocess` ENV variable for each HTTP connection
* Only lets processes run for set duration i.e. 500ms, 2s, 3s.
* Language/binding independent - can invoke any UNIX process, including built-ins such as `wc` or `cat`
* Can also run Windows native binaries

## 2. Future items

Most items are detailed [via Github issues](https://github.com/alexellis/faas/issues).

Native support is available for Docker Swarm and Kubernetes using primitive API objects in each orchestration platform.

Must have

* Developer Cloud guides:
 * for Digital Ocean
 * for Packet
 * for AWS

* Re-branding to OpenFaaS (in-progress)
 * New logo - graphic icon and text (in-progress)
 * Website / landing page (in progress)
 
* Developer guide for your first Node.js function
* Developer guide to using functions together - via pipes on client, or a "director" function on server

Should have

* helm chart
* Certifier for third-party integrations (via e2e tests) 
* AfterBurn - fork once, use many which removes almost all runtime latency - (in progress)
* Kafka queue worker implementation (async currently available by NATS Streaming)
* Non-root Docker templates for the CLI (in progress)
* Community templates for the FaaS-CLI (in progress)
* Our own "timer" (aka cron) for invoking functions on a regular basis - or a guide for setting this up via Jenkins or CRON
* Integration with a reverse proxy - such as Traefik or Kong
 * I.e. for TLS termination
 * Basic auth for /system endpoints (probably via reverse proxy)
* AARCH64 (64-bit ARM) port (dependent on Docker release schedule)
* Healthchecks for functions deployed on Kubernetes

Could have

* Multi-tenancy (in-progress for Kubernetes and Docker Swarm)
* Progress animation for building Docker images via CLI
* Built-in Docker registry with default configuration
* Docker image builder (remote service)
* Function store - list of useful predefined functions
* Supporting request parameters via route
* Configurable memory limits via "new function" pop-up (already supported by Docker compose stack)
* Scale to zero 0/0 replicas
* Guide/proxy for Flask in a function


Nice to have

* Guide for basic authentication over HTTPs (set up externally through NGinx etc)
* CRIU - (Checkpoint/Restore In Userspace) for starting serverless tasks with a high start-up cost/latency.
* Deeper tests coverage and integration tests
* Serverless Inc framework support - as a "provider"

On-going integrations in addition to Swarm and K8s:

* ECS - via Huawei
* Nomad via Hashicorp
* Hyper.sh via Hyper
* Cattle / Rancher by community

Internal research is also being done for the ACI / K8s-connector.

### Contributing

Please see [CONTRIBUTING.md](https://github.com/alexellis/faas/blob/master/CONTRIBUTING.md).
