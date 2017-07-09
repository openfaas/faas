# Roadmap

## 1. Current features

### The API Gateway

This container acts in a similar way to the API Gateway on AWS. Requests can be made to this endpoint with a JSON body.

Completed features:

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

* Asynchronous / long-running tasks

Should have

* Kubernetes support
* AARCH64 (64-bit ARM) port
* Basic auth for /system endpoints (probably via reverse proxy)

Could have

* Function store - list of useful predefined functions
* Supporting request parameters
* Configurable memory limits via "new function" pop-up (already supported by Docker compose stack)

Nice to have

* Guide for termination through NGinx or built-in TLS termination
* Guide for basic authentication over HTTPs (set up externally through NGinx etc)
* CRIU - (Checkpoint/Restore In Userspace) for warm-loading serverless tasks with a high start-up cost/latency.
* Deeper tests coverage and integration tests

## 3. Development and Contributing

If you would like to consume the project with your own functions then you can use the public images and the supplied `docker stack` file as a template (docker-compose.yml)

### Contributing

Here are a few guidelines for contributing:

* If you have found a bug please raise an issue and fill out the whole template.
* If you would like to contribute to the codebase please raise an issue to propose the change and fill out the whole template.
* If the documentation can be improved / translated etc please raise an issue to discuss. PRs for changing one or two typos aren't necessary.

**Practical stuff**

* Please sign your commits with `git commit -s` so that commits are traceable.
* Please always provide a summary of what you changed, how you did it and how it can be tested

### License

This project is licensed under the MIT License.
