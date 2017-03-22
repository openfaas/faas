# Roadmap

## 1. Current items

### The API Gateway

This container acts in a similar way to the API Gateway on AWS. Requests can be made to this endpoint with a JSON body.

Features:

* UI for viewing and testing functions deployed through stack
* Auto-scaling of replicas as load increases
* Backing off of replicas as load reduces
* Unique URL routes for serverless functions
* Instrumentation via Prometheus metrics at GET /metrics
* Bundled Prometheus stack with AlertManager
* UI enhancements to create new function through a form


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

* Asynchronous / long-running tasks
* Built-in TLS termination or guide for termination through NGinx etc
* Deeper tests coverage and integration tests
* Documentation about Alexa sample function
* Supporting request parameters

## 3. Development and Contributing

If you would like to consume the project with your own functions then you can use the public images and the supplied `docker stack` file as a template (docker-compose.yml)

### License

This project is licensed under the MIT License.

## Contributing

Here are a few guidelines for contributing:

* If you have found a bug please raise an issue.
* If the documentation can be improved / translated etc please raise an issue to discuss.
* If you would like to contribute to the codebase please raise an issue to propose the change.

> Please provide a summary of what you changed, how you did it and how it can be tested.
