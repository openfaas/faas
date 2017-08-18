## Functions as a Service (OpenFaaS)

[![Build
Status](https://travis-ci.org/alexellis/faas.svg?branch=master)](https://travis-ci.org/alexellis/faas)

![https://blog.alexellis.io/content/images/2017/08/faas_side.png](https://blog.alexellis.io/content/images/2017/08/faas_side.png)

FaaS is a framework for building serverless functions with Docker which has first class support for metrics. Any process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

**Highlights**

* Ease of use through UI portal and *one-click* install
* Write functions in any language for Linux or Windows and package in Docker/OCI image format
* Portable - runs on existing hardware or public/private cloud - [Kubernetes](https://github.com/alexellis/faas-netes) or Docker Swarm
* [CLI](http://github.com/alexellis/faas-cli) available with YAML format for templating and defining functions
* Auto-scales as demand increases

## Overview of OpenFaaS

![Stack](https://pbs.twimg.com/media/DFrkF4NXoAAJwN2.jpg)

### Function Watchdog

* You can make any Docker image into a serverless function by adding the *Function Watchdog* (a tiny Golang HTTP server)
* The *Function Watchdog* is the entrypoint allowing HTTP requests to be forwarded to the target process via STDIN. The response is sent back to the caller by writing to STDOUT from your application.

### Gateway

* The API Gateway provides an external route into your functions and collects Cloud Native metrics through Prometheus.
* Your API Gateway will scale functions according to demand by altering the service replica count in the Docker Swarm or Kubernetes API.
* A UI is baked in allowing you to invoke functions in your browser and create new ones as needed.

### CLI

Any container or process in a Docker container can be a serverless function in FaaS. Using the [FaaS CLI](http://github.com/alexellis/faas-cli) you can deploy your functions or quickly create new functions from templates such as Node.js or Python.

## Notable mentions

### Closing Keynote at Dockercon 2017

Functions as a Service or FaaS was a winner in the Cool Hacks contest for Dockercon 2017.

* [Watch my FaaS keynote at Dockercon 2017](https://blog.docker.com/2017/04/dockercon-2017-mobys-cool-hack-sessions/)

If you'd like to find the functions I used in the demos head over to the [faas-dockercon](https://github.com/alexellis/faas-dockercon/) repository.

### InfoWorld

Serdar Yegulalp	Senior Technical Writer covered FaaS in a write-up looking at serverless in the open-source world:

[Open source project uses Docker for serverless computing](http://www.infoworld.com/article/3184757/open-source-tools/open-source-project-uses-docker-for-serverless-computing.html#tk.twt_ifw)

### Online community

There is also a community being built around FaaS with talks, demos and sample functions being built out.

[Find out about community activity](https://github.com/alexellis/faas/blob/master/community.md)

## Getting started

**Jump straight in**

Run your first function in 10-15 minutes with this *new* guide.

[Your first serverless Python function with OpenFaaS](https://blog.alexellis.io/introducing-functions-as-a-service/)

**Introducing OpenFaaS blog post**

Read up on the background, the top 3 highlighted features and what's coming next:

[Introducing Functions as a Service (FaaS)](https://blog.alexellis.io/introducing-functions-as-a-service/)

**Test Drive OpenFaaS**

You can [TestDrive](https://github.com/alexellis/faas/blob/master/TestDrive.md) FaaS on your laptop in 60 seconds, or deploy to a free online Docker playground. Find out more in the [TestDrive](https://github.com/alexellis/faas/blob/master/TestDrive.md).

You can dive straight into the [sample functions here](https://github.com/alexellis/faas/blob/master/sample-functions/README.md). You'll find hello-world examples for the most common programming languages including: Golang, DotNet Core, Java, NodeJS, Python even BusyBox.

The [faas-cli](https://github.com/alexellis/faas-cli/) lets you speed up development by creating functions from templates for:

* [Node.js](https://github.com/alexellis/faas-cli/blob/master/test_node.sh)
* [Python](https://github.com/alexellis/faas-cli/blob/master/test_python.sh)
* Ruby
* CSharp

..or whatever langauge you can create your own template for.

You can also use the CLI to deploy to your own FaaS API Gateway with a single command.

**Contribute to OpenFaaS**

FaaS enables you to run your serverless functions in whatever language you like, wherever you like - for however long you need.

Contributions to the project are welcome - please send in issues and questions through Github.

[See issues and PRs](https://github.com/alexellis/faas/issues)

*What about the name?*

FaaS is becoming OpenFaas, see more here: [Issue 123](https://github.com/alexellis/faas/issues/123)

