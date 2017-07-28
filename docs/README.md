## Functions as a Service (FaaS)

FaaS is a framework for building serverless functions on Docker with first class support for metrics. Any UNIX process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

## Concept

* Each container has a watchdog process that hosts a web server allowing a post request to be forwarded to a desired process via STDIN. The response is sent back to the caller via STDOUT.

* The API Gateway provides an external route into your functions and collects metrics in Prometheus. The gateway will scale functions according to demand by mangaging Docker Swarm replicas as throughput increases. A UI is baked in allowing you to invoke functions in your browser and create new ones as needed.

## FaaS Stack

FaaS is an open-source project written in Golang and licensed under the MIT license.

![Stack](http://blog.alexellis.io/content/images/2017/04/faas_hi.png)

**Highlights:**

* Ease of use through UI portal
* Setup a working environment with one script
* Portable - runs on any hardware supported by Docker

* Any process that can run in Docker can be a serverless function

* Baked-in Prometheus metrics and logging
* Auto-scales as demand increases

## Notable mentions

### Closing Keynote at Dockercon 2017

Functions as a Service or FaaS was a winner in the Cool Hacks contest for Dockercon 2017.

* [Watch my FaaS keynote at Dockercon 2017](https://blog.docker.com/2017/04/dockercon-2017-mobys-cool-hack-sessions/)

If you'd like to find the functions I used in the demos head over to the [faas-dockercon](https://github.com/alexellis/faas-dockercon/) repository.

### InfoWorld

Serdar Yegulalp	Senior Technical Writer covered FaaS in a write-up looking at serverless in the open-source world:

[Open source project uses Docker for serverless computing](http://www.infoworld.com/article/3184757/open-source-tools/open-source-project-uses-docker-for-serverless-computing.html#tk.twt_ifw)

### Community activity

There is also a community being built around FaaS with talks, demos and sample functions being built out.

[Find out about community activity](https://github.com/alexellis/faas/blob/master/community.md)

## Getting started

**Test Drive FaaS**

You can [TestDrive](https://github.com/alexellis/faas/blob/master/TestDrive.md) FaaS on your laptop in 60 seconds, or deploy to a free online Docker playground. Find out more in the [TestDrive](https://github.com/alexellis/faas/blob/master/TestDrive.md).

You can dive straight into the [sample functions here](https://github.com/alexellis/faas/blob/master/sample-functions/README.md). You'll find hello-world examples for the most common programming languages including: Golang, DotNet Core, Java, NodeJS, Python even BusyBox.

There is even CLI called [faas-cli](https://github.com/alexellis/faas-cli/) which lets you speed up development by creating functions from templates for [Node.js](https://github.com/alexellis/faas-cli/blob/master/test_node.sh) or [Python](https://github.com/alexellis/faas-cli/blob/master/test_python.sh). You can also use the CLI to deploy to your own FaaS API Gateway with a single command.

**Contribute to FaaS**

FaaS enables you to run your serverless functions in whatever language you like, wherever you like - for however long you need.

Contributions to the project are welcome - please send in issues and questions through Github.

[See issues and PRs](https://github.com/alexellis/faas/issues)

