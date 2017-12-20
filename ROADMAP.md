# Roadmap

## 1. Current features

For an overview of features in August 2017 read the following post:

* [Introducing OpenFaaS (Functions as a Service)](https://blog.alexellis.io/introducing-functions-as-a-service/)

For the latest updates see [blog.alexellis.io/tag/openfaas](https://blog.alexellis.io/tag/openfaas/)

## Primary GitHub Organisation:

* https://github.com/openfaas/faas
* https://github.com/openfaas/faas-netes
* https://github.com/openfaas/faas-cli
* https://github.com/openfaas/nats-queue-worker

## Incubator GitHub Organisation:

* https://github.com/openfaas-incubator

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
* helm chart
* Developer guide to using functions together - via pipes on client, or a "director" function on server
* Developer guide for DigitalOcean
* Re-branding to OpenFaaS
 * New logo - graphic icon and text
 * Website / landing page
* Non-root Docker templates for the CLI
* Integration with a reverse proxy - such as Traefik or Kong
 * I.e. for TLS termination
 * Basic auth for /system endpoints (probably via reverse proxy)
* AARCH64 (64-bit ARM) port (dependent on Docker release schedule)
* Healthchecks for functions deployed on Kubernetes
* Supporting request parameters via query-string
* Custom end-to-end label support for functions
* Supporting request parameters via route - QueryString / Path
* Guide for basic authentication over HTTPs (set up externally through NGinx etc)

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

## 2. Future/current work (subject to change)

If you need an up-to-date picture of what is current and ready-for-use please reach out to the OpenFaaS maintainers through our Slack comunity. Most items are detailed [via Github issues](https://github.com/openfaas/faas/issues).

Native support is available for Docker Swarm and Kubernetes using primitive API objects in each orchestration platform.

Must have

* Dedicated blog site built from GitHub pages or Hugo
* Developer Cloud guides:
 * for Digital Ocean (done)
 * for Packet
 * for AWS
* Developer guide for your first Node.js function

* Configurable memory limits per function exposed through API
* Certifier for third-party integrations - via e2e tests (in progress) 
* Return stderr or link to stderr via function invocation (For VMWare / Mark Peek)
* Community function templates for the FaaS-CLI (in progress)

Should have

* Docker image builder as remote service based upon Moby
* Kafka-Connector for the API Gateway
* AfterBurn - fork once, use many which removes almost all runtime latency - (Alpha available)
* Kafka queue worker implementation (async currently available by NATS Streaming) - available in pending PR

* Our own "timer" (aka cron) for invoking functions on a regular basis - or a guide for setting this up via Jenkins or CRON
* Guide/proxy for Flask in a function


Could have

* Multi-tenancy (in-progress for Kubernetes and Docker Swarm)
* Progress animation for building Docker images via CLI
* Built-in Docker registry with default configuration

* Function store - list of useful predefined functions
* Configurable memory limits via "new function" pop-up (already supported by Docker compose stack)
* Scale to zero 0/0 replicas
* Serverless Inc framework support - as a "provider" (in progress)

Nice to have

* CRIU - (Checkpoint/Restore In Userspace) for starting serverless tasks with a high start-up cost/latency.
* Deeper tests coverage and integration tests

On-going integrations in addition to Swarm and K8s:

* ECS - via Huawei
* Nomad via Hashicorp / Nic Jackson
* Hyper.sh via Hyper
* Cattle / Rancher by Ken in the community
* DC/OS via Alberto in the community

Internal research is also being done for the ACI / AKS / K8s-connector.

### Contributing

Please see [CONTRIBUTING.md](https://github.com/openfaas/faas/blob/master/CONTRIBUTING.md).
