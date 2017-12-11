# Gateway

The API Gateway provides an external route into your functions and collects Cloud Native metrics through Prometheus, as well as a UI for creating and invoking functions. 
The gateway will scale functions according to demand by altering the service replica count in the Docker Swarm or Kubernetes API.

Swagger docs: https://github.com/openfaas/faas/tree/master/api-docs

## Logs

Logs are available at the function level and can be accessed through Swarm or Kubernetes using native tooling. You can also install a Docker logging driver to aggregate your logs. By default functions will not write the request and response bodies to stdout. You can toggle this behaviour by setting `read_debug` for the request and `write_debug` for the response.

## Tracing

An "X-Call-Id" header is applied to every incoming call through the gateway and is usable for tracing and monitoring calls. We use a UUID for this string.

Header:

```
X-Call-Id
```

Within a function this is available as `Http_X_Call_Id`.

## Environmental overrides
The gateway can be configured through the following environment variables: 

| Option                 | Usage             |
|------------------------|--------------|
| `write_timeout`        | HTTP timeout for writing a response body from your function (in seconds). Default: `8`  |
| `read_timeout`         | HTTP timeout for reading the payload from the client caller (in seconds). Default: `8` |
| `functions_provider_url`             | URL of an alternate microservice to manage functions (e.g., Kubernetes). When given, this overrides the default Docker Swarm provider.  |
| `faas_nats_address`          | Address of NATS service. Required for asynchronous mode. |
| `faas_nats_port`    | Port for NATS service. Requrired for asynchronous mode. |
| `faas_prometheus_host`         | Host to connect to Prometheus. Default: `"prometheus"`.  |
| `faas_promethus_port`         | Port to connect to Prometheus. Default: `9090`. |
