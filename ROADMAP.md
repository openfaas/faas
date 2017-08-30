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
* Kubernetes support

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

## 2. Future items

Most items are detailed [via Github issues](https://github.com/alexellis/faas/issues).

Must have

* Re-branding to OpenFaaS
 * New logo - graphic icon and text (in progress)
 * Website / landing page (in progress)
* Asynchronous / long-running tasks (PR in testing)

Should have

* AARCH64 (64-bit ARM) port (dependent on Docker release schedule)
* Integration with a reverse proxy - such as Traefik or Kong
 * Basic auth for /system endpoints (probably via reverse proxy)
* CLI - list functions / query function info
* OS constraints in the deploy function API
* Healthchecks for functions deployed on Kubernetes 

Could have

* Built-in Docker registry with default configuration
* Docker image builder (remote service)
* Function store - list of useful predefined functions
* Supporting request parameters
* Configurable memory limits via "new function" pop-up (already supported by Docker compose stack)

Nice to have

* Developer Cloud guide:
 * for Digital Ocean
 * for Packet
 
* Developer guide for your first Node.js function
* Developer guide to using functions together - via pipes on client, or a "director" function on server

* Documentation on using CRON / JenkinsCI for invoking functions on a timed basis

* Guide for termination through NGinx or built-in TLS termination
* Guide for basic authentication over HTTPs (set up externally through NGinx etc)
* CRIU - (Checkpoint/Restore In Userspace) for warm-loading serverless tasks with a high start-up cost/latency.
* Deeper tests coverage and integration tests

### Contributing

Please see [CONTRIBUTING.md](https://github.com/alexellis/faas/blob/master/CONTRIBUTING.md).
