# faas - Functions As A Service

This project provides a way to run Docker containers as functions on Swarm Mode. 

* Each container has a watchdog process that hosts a web server allowing a JSON post request to be forwarded to a desired process via STDIN. The respose is sent to the caller via STDOUT.
* A gateway provides a view to the containers/functions to the public Internet and collects metrics for Prometheus and in a future version will manage replicas and scale as throughput increases.

## Quickstart

Minimum requirements: 

* Docker 1.13-RC (to support attachable overlay networks)
* At least a single host in Swarm Mode. (run `docker swarm init`)

Check your `docker version` and upgrade to one of the latest 1.13-RCs from the [Docker Releases page](https://github.com/docker/docker/releases). This is already available through the Beta channel in Docker for Mac.

#### Create an attachable network for the gateway and functions to join

```
# docker network create --driver overlay --attachable functions
```

#### Start the gateway

```
# docker pull alexellisio/faas-gateway:latest
# docker rm -f gateway;
# docker run -d -v /var/run/docker.sock:/var/run/docker.sock --name gateway -p 8080:8080 \
  --network=functions alexellisio/faas-gateway:latest
```

#### Start at least one of the serverless functions:

Here we start an echo service using the `cat` command found in a shell.

```
# docker service rm catservice
# docker service create --network=functions --name catservice alexellisio/faas-catservice:latest
```

#### Now send an event to the API gateway

* Method 1 - use the service name as a URL:

```
# curl -X POST --data-binary @$HOME/.ssh/known_hosts -v http://localhost:8080/function/catservice
```

* Method 2 - use the X-Function header:

```
# curl -X POST -H 'x-function: catservice' --data-binary @$HOME/.ssh/known_hosts -v http://localhost:8080/
```

#### Build your own function

Visit the accompanying blog post to find out how to build your own function in whatever programming language you prefer.

[FaaS blog post](http://blog.alexellis.io/functions-as-a-service/)

# Overview

## the gateway

This container acts in a similar way to the API Gateway on AWS. Requests can be made to this endpoint with a JSON body.

**Incoming requests and routing**

There are three options for routing:

* Functions created on the overlay network can be invoked by: http://localhost:8080/function/{servicename}
* Routing automatically detects Alexa SDK requests and forwards to a service name (function) that matches the Intent name
* Routing is enabled through a `X-Function` header which matches a service name (function) directly.

Features:

* [todo] auto-scaling of replicas as load increases
* [todo] backing off of replicas as load reduces
* [todo] unique URL routes for serverless functions
* instrumentation via Prometheus metrics at GET /metrics

## the watchdog


This binary fwatchdog acts as a watchdog for your function. Features:

* Static binary in Go
* Listens to HTTP requests over swarm overlay network
* Spawns process set in `fprocess` ENV variable for each HTTP connection
* [todo] Only lets processes run for set duration i.e. 500ms, 2s, 3s.
* Language/binding independent

### Additional technical debt:

* Must switch over to timeouts for HTTP.post via HttpClient.
* Coverage with unit tests
* Update quick-start to use `docker stack deploy`

## Building a development environment:

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

## Prometheus metrics / instrumentation

Standard go metrics and function invocation count / duration are available at http://localhost:8080/metrics/
