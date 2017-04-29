## Functions as a Service (FaaS)

FaaS is a framework for building serverless functions on Docker with first class support for metrics. Any UNIX process can be packaged as a function enabling you to consume a range of web events without repetitive boiler-plate coding.

## Concept

* Each container has a watchdog process that hosts a web server allowing a post request to be forwarded to a desired process via STDIN. The response is sent back to the caller via STDOUT.

* The API Gateway provides an external route into your functions and collects metrics in Prometheus. The gateway will scale functions according to demand by mangaging Docker Swarm replicas as throughput increases. A UI is baked in allowing you to invoke functions in your browser and create new ones as needed.

## Closing Keynote at Dockercon 2017

Functions as a Service or FaaS was a winner in the Cool Hacks contest for Dockercon 2017.

* [Watch my FaaS keynote at Dockercon 2017](https://blog.docker.com/2017/04/dockercon-2017-mobys-cool-hack-sessions/)

If you'd like to find the functions I used in the demos head over to the [faas-dockercon](https://github.com/alexellis/faas-dockercon/) repository.
