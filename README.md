# faas
Functions as a service

Minimum requirements: Docker 1.13

gateway
=======

This container acts in a similar way to the API Gateway on AWS. Requests can be made to this endpoint with a JSON body.

Features:

* auto-scaling of replicas as load increases
* backing off of replicas as load reduces
* unique URL routes for serverless functions
* instrumentation via Prometheus metrics at GET /metrics

watchdog
========

This binary fwatchdog acts as a watchdog for your function. Features:

* Static binary in Go
* Listens to HTTP requests over swarm overlay network
* Spawns process set in `fprocess` ENV variable for each HTTP connection
* Only lets processes run for set duration i.e. 500ms, 2s, 3s.
* Language/binding independent

Complete example:
=================

```
# docker network create --driver overlay --attachable functions
# git clone https://github.com/alexellis/faas && cd faas
# cd watchdog
# ./build.sh
# docker build -t catservice .
# docker service rm catservice ; docker service create --network=functions --name catservice catservice
# cd ..
# cd gateway
# docker build -t server . ;docker rm -f server; docker run -v /var/run/docker.sock:/var/run/docker.sock --name server -p 8080:8080 --network=functions server
```

Accessing the `cat` (read echo) service:

```
# curl -X POST -H 'x-function: catservice' --data-binary @/etc/hostname -v http://localhost:8080/curl -X POST -H 'x-function: catservice' --data-binary @$HOME/.ssh/known_hosts -v http://localhost:8080/
```

