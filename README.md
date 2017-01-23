# faas - Functions As A Service

FaaS is a platform for building serverless functions on Docker Swarm Mode with first class metrics. Any UNIX process can be packaged as a function in FaaS enabling you to consume a range of web events without repetitive boiler-plate coding.

## Concept

* Each container has a watchdog process that hosts a web server allowing a JSON post request to be forwarded to a desired process via STDIN. The respose is sent to the caller via STDOUT.
* A gateway provides a view to the containers/functions to the public Internet and collects metrics for Prometheus and in a future version will manage replicas and scale as throughput increases.

## Minimum requirements: 
* Docker 1.13-RC (to support attachable overlay networks)
* At least a single host in Swarm Mode. (run `docker swarm init`)

> For more information on Swarm mode and configuration please read the [Swarm Mode tutorial](https://docs.docker.com/engine/swarm/swarm-tutorial/).

Check your `docker version` and upgrade to one of the latest 1.13-RCs from the [Docker Releases page](https://github.com/docker/docker/releases). This is already available through the Beta channel in Docker for Mac.

## Quickstart with `docker stack deploy`

For a complete stack of Prometheus, the gateway and the DockerHubStats function: 

* Simply run `./deploy_stack.sh` - following that you can find out information about the services like this:

```
# docker stack ls
NAME  SERVICES
func  3

# docker stack ps func
ID            NAME               IMAGE                                  NODE  DESIRED STATE  CURRENT STATE         
rhzej73haufd  func_gateway.1     alexellis2/faas-gateway:latest         moby  Running        Running 26 minutes ago
fssz6unq3e74  func_hubstats.1    alexellis2/faas-dockerhubstats:latest  moby  Running        Running 27 minutes ago
nnlzo6u3pilg  func_prometheus.1  quay.io/prometheus/prometheus:latest   moby  Running        Running 27 minutes ago
```

* Then head over to http://localhost:9090 for your Prometheus metrics

* Your function can be accessed via the gateway like this:

**Sample function: Docker Hub Stats (hubstats)**

```
# curl -X POST http://localhost:8080/function/func_hubstats -d "alexellis2"
The organisation or user alexellis2 has 99 repositories on the Docker hub.

# curl -X POST http://localhost:8080/function/func_hubstats -d "library"
The organisation or user library has 128 repositories on the Docker hub.
```

The `-d` value passes in the argument for your function. This is read via STDIN and used to query the Docker Hub to see how many images you've created/pushed.

**Sample function: webhook stasher (webhookstash)**

Another cool sample function is the Webhook Stasher which saves the body of any data posted to the service to the container's filesystem. Each file is written with the filename of the UNIX time.

```
# curl -X POST http://localhost:8080/function/func_webhookstash -d '{"event": "fork", "repo": "alexellis2/faas"}'
Webhook stashed

# docker ps|grep stash
d769ca70729d        alexellis2/faas-webhookstash@sha256:b378f1a144202baa8fb008f2c8896f5a8

# docker exec d769ca70729d find .
.
./1483999546817280727.txt
./1483999702130689245.txt
./1483999702533419188.txt
./1483999702978454621.txt
./1483999703284879767.txt
./1483999719981227578.txt
./1483999720296180414.txt
./1483999720666705381.txt
./1483999720961054638.txt
```

**Sample function: Node OS Info (nodeinfo)**

Grab OS, CPU and other info via a Node.js container using the `os` module.

```
# curl -X POST http://localhost:8080/function/func_nodeinfo -d ''

linux x64 [ { model: 'Intel(R) Xeon(R) CPU E5-2670 v2 @ 2.50GHz',
    speed: 2500,
    times: 
     { user: 3754430800,
       nice: 2450200,
       sys: 885352200,
       idle: 25599742200,
       irq: 0 } },
...
```

> Why not start the code on play-with-docker.com and then configure a Github repository to send webhook to the function?

If you're looking for a UI checkout the [Postman plugin for Chrome](https://www.getpostman.com) where you can send POSTs without needing `curl`.

## Manual quickstart

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

#### Setup a Prometheus instance

Please review the following quickstart example and edit the configuration in `monitor/prometheus.yml`.

* [Quickstart-Prometheus](https://github.com/alexellis/quickstart-prometheus)

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
 * Have `docker stack deploy` include a pre-configured Prometheus instance.

## Development

For development of the FaaS framework / library read on. If you would like to consume the project with your own functions then you can use the public images and the supplied `docker stack` file as a template (docker-compose.yml)

### Contributing

* If you have found a bug please raise an issue.
* If the documentation can be improved / translated etc please raise an issue to discuss.
* If you would like to contribute to the codebase please raise an issue to propose the change.

### Building a development environment:

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
