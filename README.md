## Functions As A Service (faas)

FaaS is a framework for building serverless functions on Docker Swarm with first class support for metrics. Any UNIX process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

## Concept

* Each container has a watchdog process that hosts a web server allowing a post request to be forwarded to a desired process via STDIN. The response is sent back to the caller via STDOUT.

* The API Gateway provides an external route into your functions and collects metrics in Prometheus. The gateway will scale functions according to demand by mangaging Docker Swarm replicas as throughput increases. A UI is baked in allowing you to invoke functions in your browser and create new ones as needed.

> ### [Read the story of FaaS on my blog](http://blog.alexellis.io/functions-as-a-service/) or find out more about the project below.

[![Build
Status](https://travis-ci.org/alexellis/faas.svg?branch=master)](https://travis-ci.org/alexellis/faas)

## Minimum requirements: 
* Docker 1.13 (to support Docker Stacks)
* At least a single host in Docker Swarm Mode. (run `docker swarm init`)

## TestDrive

A one-line script is provided to help you get started quickly. You can test-drive FaaS with a set of sample functions as defined in the provided [docker-compose.yml](https://github.com/alexellis/faas/blob/master/docker-compose.yml) file. 

Use your own laptop or the free community-run Docker playground: play-with-docker.com.

**Highlights:**

* Ease of use through UI portal
* Setup a working environment with one script
* Portable - runs on any hardware

* Baked-in Prometheus metrics
* Any container can be a function
* Auto-scales as demand increases

### [Begin the TestDrive](https://github.com/alexellis/faas/blob/master/TestDrive.md)

Here is a screenshot of the API gateway portal - designed for ease of use.

![Portal](https://pbs.twimg.com/media/C7bkpZbWwAAnKsx.jpg)

### Ongoing development/screenshots:

FaaS is still expanding and growing, check out the developments around:

* [Auto-scaling through Prometheus alerts](https://twitter.com/alexellisuk/status/825295438412709888)
* [Prometheus alert example](https://twitter.com/alexellisuk/status/823262200236277762)
* [Invoke functions through UI](https://twitter.com/alexellisuk/status/823262200236277762)
* [Create new functions through UI](https://twitter.com/alexellisuk/status/835047437588905984)
* [Various sample functions](https://github.com/alexellis/faas/blob/master/docker-compose.yml)

* [ARM / Raspberry Pi support](https://gist.github.com/alexellis/665332cd8bd9657c9649d0cd6c2dc187)

## Package existing code as functions

* [Package your function](https://github.com/alexellis/faas/blob/master/DEV.md)

## Roadmap and contributing

* [Read the Roadmap](https://github.com/alexellis/faas/blob/master/ROADMAP.md)

## Additional content

#### Would you prefer a video overview?

See how to deploy FaaS onto play-with-docker.com and Docker Swarm in 1-2 minutes. See the sample functions in action and watch the graphs in Prometheus as we ramp up the amount of requests. 

* [Deep Dive into Functions as a Service (FaaS) on Docker](https://www.youtube.com/watch?v=sp1B7l5mEzc)

#### Prometheus metrics are built-in

Prometheus is built into FaaS and the sample stack, so you can check throughput for each function individually with a rate function in the UI at port 9090 on your Swarm manager.

If you are new to Prometheus, you can start learning about metrics and monitoring on my blog:

> [Monitor your applications with Prometheus](http://blog.alexellis.io/prometheus-monitoring/)

![Prometheus UI](https://pbs.twimg.com/media/C7bkiT9X0AASVuu.jpg)
