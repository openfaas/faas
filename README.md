## Functions as a Service (FaaS)

[![Build
Status](https://travis-ci.org/alexellis/faas.svg?branch=master)](https://travis-ci.org/alexellis/faas)

FaaS is a framework for building serverless functions on Docker with first class support for metrics. Any UNIX process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

**Highlights**

* Ease of use through UI portal and *one-click* install
* Write functions in any language for Linux or Windows
* Portable - runs on your own hardware, or any cloud
* Auto-scales as demand increases

## Concept

* Each container has a watchdog process that hosts a web server allowing a post request to be forwarded to a desired process via STDIN. The response is sent back to the caller via STDOUT.

* The API Gateway provides an external route into your functions and collects metrics in Prometheus. The gateway will scale functions according to demand by managing Docker Swarm replicas as throughput increases. A UI is baked in allowing you to invoke functions in your browser and create new ones as needed.

![Stack](http://blog.alexellis.io/content/images/2017/04/faas_hi.png)

## Closing Keynote at Dockercon 2017

Functions as a Service or FaaS was a winner in the Cool Hacks contest for Dockercon 2017.

* [Watch my FaaS keynote at Dockercon 2017](https://blog.docker.com/2017/04/dockercon-2017-mobys-cool-hack-sessions/)

If you'd like to find the functions I used in the demos head over to the [faas-dockercon](https://github.com/alexellis/faas-dockercon/) repository.

**Background story**

This is my original blog post on FaaS from Janurary: [Functions as a Service blog post](http://blog.alexellis.io/functions-as-a-service/)

## TestDrive

A one-line script is provided to help you get started quickly. You can test-drive FaaS with a set of sample functions as defined in the provided [docker-compose.yml](https://github.com/alexellis/faas/blob/master/docker-compose.yml) file. 

Use your own laptop or the free community-run Docker playground: play-with-docker.com.

[![Try in PWD](https://cdn.rawgit.com/play-with-docker/stacks/cff22438/assets/images/button.png)](http://play-with-docker.com?stack=https://raw.githubusercontent.com/alexellis/faas/master/docker-compose.yml&stack_name=func)

### [Begin the TestDrive](https://github.com/alexellis/faas/blob/master/TestDrive.md)

Here is a screenshot of the API gateway portal - designed for ease of use.

![Portal](https://pbs.twimg.com/media/C7bkpZbWwAAnKsx.jpg)

### Community

* [Read blogs/articles and find events about FaaS](https://github.com/alexellis/faas/blob/master/community.md)

## Package existing code as functions

* [Package your function](https://github.com/alexellis/faas/blob/master/DEV.md)
* [Experimental CLI for templating/deploying functions](https://github.com/alexellis/faas-cli)

## Roadmap and contributing

* [Read the Roadmap](https://github.com/alexellis/faas/blob/master/ROADMAP.md)

### Ongoing development/screenshots:

FaaS is still expanding and growing, check out the developments around:

* [Auto-scaling through Prometheus alerts](https://twitter.com/alexellisuk/status/825295438412709888)
* [Prometheus alert example](https://twitter.com/alexellisuk/status/823262200236277762)
* [Invoke functions through UI](https://twitter.com/alexellisuk/status/823262200236277762)
* [Create new functions through UI](https://twitter.com/alexellisuk/status/835047437588905984)
* [Various sample functions](https://github.com/alexellis/faas/blob/master/docker-compose.yml)
* [Integration between IFTTT and Slack](https://twitter.com/alexellisuk/status/857300138745876483)
* [Mixed Linux / Windows serverless functions](https://github.com/alexellis/faas/pull/79)
* [ARM / Raspberry Pi support](https://gist.github.com/alexellis/665332cd8bd9657c9649d0cd6c2dc187)

## Minimum requirements: 
* Docker 1.13 (to support Docker Stacks)
* At least a single host in Docker Swarm Mode. (run `docker swarm init`)

## Additional content

#### Would you prefer a video overview?

**Dockercon closing keynote - Cool Hacks demos with Alexa/Github - April 19th 2017**

[Watch on YouTube](https://www.youtube.com/watch?v=-h2VTE9WnZs&t=961s&list=PLlIapFDp305AiwA17mUNtgi5-u23eHm5j&index=1)

**FaaS tour of features including an Alexa Skill - April 9th 2017**

[Watch on YouTube](https://www.youtube.com/watch?v=BK076ChLKKE)

**FaaS deep dive - 11th Jan 2017**

See how to deploy FaaS onto play-with-docker.com and Docker Swarm in 1-2 minutes. See the sample functions in action and watch the graphs in Prometheus as we ramp up the amount of requests. 

* [Deep Dive into Functions as a Service (FaaS) on Docker](https://www.youtube.com/watch?v=sp1B7l5mEzc)

#### Prometheus metrics are built-in

Prometheus is built into FaaS and the sample stack, so you can check throughput for each function individually with a rate function in the UI at port 9090 on your Swarm manager.

If you are new to Prometheus, you can start learning about metrics and monitoring on my blog:

> [Monitor your applications with Prometheus](http://blog.alexellis.io/prometheus-monitoring/)

![Prometheus UI](https://pbs.twimg.com/media/C7bkiT9X0AASVuu.jpg)

You can also link this to a Grafana dashboard and see auto-scaling live in action:

![](https://pbs.twimg.com/media/C9caE6CXUAAX_64.jpg:large)

Sample dashboard JSON file available [here](https://github.com/alexellis/faas/blob/master/contrib/grafana.json)
