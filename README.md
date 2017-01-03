# faas - Functions As A Service

This project provides a way to run Docker containers as functions on Swarm Mode. 

* Each container has a watchdog process that hosts a web server allowing a JSON post request to be forwarded to a desired process via STDIN. The respose is sent to the caller via STDOUT.
* A gateway provides a view to the containers/functions to the public Internet and collects metrics for Prometheus and in a future version will manage replicas and scale as throughput increases.

Minimum requirements: Docker 1.13

gateway
=======

This container acts in a similar way to the API Gateway on AWS. Requests can be made to this endpoint with a JSON body.

**Incoming requests and routing**

There are three options for routing:

* Routing is enabled through a `X-Function` header which matches a service name (function) directly.
* Routing automatically detects Alexa SDK requests and forwards to a service name (function) that matches the Intent name
* [todo] individual routes can be set up mapping to a specific service name (function).

Features:

* [todo] auto-scaling of replicas as load increases
* [todo] backing off of replicas as load reduces
* [todo] unique URL routes for serverless functions
* instrumentation via Prometheus metrics at GET /metrics

watchdog
========

This binary fwatchdog acts as a watchdog for your function. Features:

* Static binary in Go
* Listens to HTTP requests over swarm overlay network
* Spawns process set in `fprocess` ENV variable for each HTTP connection
* [todo] Only lets processes run for set duration i.e. 500ms, 2s, 3s.
* Language/binding independent

Complete example:
=================

To use multiple hosts you should push your services (functions) to the Docker Hub or a registry accessible to all nodes.

```
# docker network create --driver overlay --attachable functions
# git clone https://github.com/alexellis/faas && cd faas
# cd watchdog
# ./build.sh
# cd ../sample-functions/catservice/
# cp ../../watchdog/fwatchdog ./
# docker build -t catservice . ; docker service rm catservice ; docker service create --network=functions --name catservice catservice
# cd ../../
# cd gateway
# docker build -t server . ;docker rm -f server; docker run -v /var/run/docker.sock:/var/run/docker.sock --name server -p 8080:8080 --network=functions server
```

Accessing the `cat` (read echo) service:

```
# curl -X POST -H 'x-function: catservice' --data-binary @$HOME/.ssh/known_hosts -v http://localhost:8080/

# curl -X POST -H 'x-function: catservice' --data-binary @/etc/hostname -v http://localhost:8080/
```

Prometheus metrics / instrumentation
====================================

* Standard go metrics and function invocation count / duration are available at http://localhost:8080/metrics/

