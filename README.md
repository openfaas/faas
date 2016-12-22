# faas
Functions as a service

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

