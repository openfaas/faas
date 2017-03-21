## faas - Functions As A Service

FaaS is a platform for building serverless functions on Docker Swarm Mode with first class metrics. Any UNIX process can be packaged as a function in FaaS enabling you to consume a range of web events without repetitive boiler-plate coding.

## Concept

* Each container has a watchdog process that hosts a web server allowing a JSON post request to be forwarded to a desired process via STDIN. The respose is sent to the caller via STDOUT.
* A gateway provides a view to the containers/functions to the public Internet and collects metrics for Prometheus and in a future version will manage replicas and scale as throughput increases.

> [Read the story on my blog](http://blog.alexellis.io/functions-as-a-service/) or find out more below.

[![Build
Status](https://travis-ci.org/alexellis/faas.svg?branch=master)](https://travis-ci.org/alexellis/faas)

## Minimum requirements: 
* Docker 1.13 (to support attachable overlay networks)
* At least a single host in Swarm Mode. (run `docker swarm init`)

## TestDrive

You can test-drive FaaS with a set of sample functions as defined in docker-compose.yml on play-with-docker.com for free, or on your own laptop.

* [Begin the TestDrive instructions](https://github.com/alexellis/faas/blob/master/TestDrive.md)

### Ongoing development/screenshots:

FaaS is still expanding and growing, check out the developments around:

* [Auto-scaling through Prometheus alerts](https://twitter.com/alexellisuk/status/825295438412709888)
* [Prometheus alert example](https://twitter.com/alexellisuk/status/823262200236277762)
* [Invoke functions through UI](https://twitter.com/alexellisuk/status/823262200236277762)
* [Create new functions through UI](https://twitter.com/alexellisuk/status/835047437588905984)
* [Various sample functions](https://github.com/alexellis/faas/blob/master/docker-compose.yml)
* [ARM / Raspberry Pi support](https://github.com/alexellis/faas/blob/master/docker-compose.armhf.yml)

## Develop your own functions

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
