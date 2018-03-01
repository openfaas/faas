## Notes on load testing

### Checklist

Performance testing should only be carried out with Kubernetes. 

* [ ] I have created a test-plan with a hypothesis and documented my method so I can share it with the project team.
* [ ] I'm using a performance testing tool such as jMeter, LoadRunner or Gattling
* [ ] My environment is hosted in an isolated and repeatable environment
* [ ] I have extended or removed memory limits / quotas
* [ ] I have created my own function using one of the new HTTP templates
* [ ] I understand the difference between the original default watchdog which forks one process per request and the new of-watchdog's HTTP mode and I am using that
* [ ] I have turned off `write_debug` and `read_debug` so that the logs for the function are kept sparse
* [ ] I am monitoring / collecting logs from the core services and function under test
* [ ] I am monitoring the system for feedback through Prometheus and / or Grafana - i.e. throughput and 200/500 errors
* [ ] ~~If running on Docker Swarm I've verified that I am using a proper HEALTHCHECK (read up more on watchdog readme)~~
* [ ] I am not using Docker Swarm
* [ ] I am using Kubernetes 1.8 or 1.9

> The [current version of OpenFaaS templates](https://github.com/openfaas/templates) use the original `watchdog` which `forks` processes - a bit like CGI. The newer watchdog [of-watchdog](https://github.com/openfaas-incubator/of-watchdog) is more similar to fastCGI/HTTP and should be used for any benchmarking or performance testing along with one of the newer templates.

of-watchdog templates:

* [Node8 HTTP template](https://github.com/openfaas-incubator/node8-express-template)
* [Golang HTTP template](https://github.com/alexellis/golang-http-template)

### Common mistakes for performance-testing a project:

* Not communicating intent

Communicate your intent to the project so that we can ensure you haven't missed anything and can share results we have obtained during our own testing. Asking arbitary questions out of context will result in a poor interaction with the community and project.

* Using an inappropriate method

There is a differnce between performance testing and Denial of Service DoS attacks. Use tools which allow a gradual ramp-up and controlled conditions such as jMeter, LoadRunner or Gattling.

* Choosing an inappropriate test environment

Do not try to performance test OpenFaaS on your laptop within a VM - this carries an overhead of virtualisation and will likely cause contention.

The test environment needs to replicate the production environment you are likely to use. Take note that most AWS virtual machines are subject to CPU throttling and a credits system which will make performance testing hard and unscientific.

* Poor choice of test function

There are several sample functions provided in the project, but that does not automatically qualify them for benchmarking or load-testing. It's important to create your own function and understand exactly what is being put into it so you can measure it efficiently.

* Ignoring CPU / memory limits

OpenFaaS enforces memory limits on core services. If you are going to perform a high load test you will want to extend these beyond the defaults or remove them completely.

