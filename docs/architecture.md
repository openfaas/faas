# Architecture

![Stack](https://pbs.twimg.com/media/DFrkF4NXoAAJwN2.jpg)

## Function Watchdog

* You can make any Docker image into a serverless function by adding the *Function Watchdog* (a tiny Golang HTTP server)
* The *Function Watchdog* is the entrypoint allowing HTTP requests to be forwarded to the target process via STDIN. The response is sent back to the caller by writing to STDOUT from your application.

## API Gateway / UI Portal

* The API Gateway provides an external route into your functions and collects Cloud Native metrics through Prometheus.
* Your API Gateway will scale functions according to demand by altering the service replica count in the Docker Swarm or Kubernetes API.
* A UI is baked in allowing you to invoke functions in your browser and create new ones as needed.

!!! note
    The API Gateway is a RESTful micro-service and you can view the [Swagger docs here](https://github.com/openfaas/faas/tree/master/api-docs).

## Prometheus/Grafana

Example of a Grafana dashboard linked to OpenFaaS showing auto-scaling live in action:

![](https://pbs.twimg.com/media/C9caE6CXUAAX_64.jpg:large)

Sample dashboard JSON file available [here](https://github.com/openfaas/faas/blob/master/contrib/grafana.json)

## CLI

Any container or process in a Docker container can be a serverless function in FaaS. Using the [FaaS CLI](http://github.com/openfaas/faas-cli) you can deploy your functions or quickly create new functions from templates such as Node.js or Python.

!!! note
    The CLI is effectively a RESTful client for the API Gateway.

When you have OpenFaaS configured you can [get started with the CLI here](https://blog.alexellis.io/quickstart-openfaas-cli/)

## Function examples

You can generate new functions using the FaaS-CLI and built-in templates or use any binary for Windows or Linux in a Docker container.

### Python

*handler.py*
```python
import requests

def handle(req):
        r =  requests.get(req, timeout = 1)
        print(req +" => " + str(r.status_code))
```

### Node.js

*handler.js*
```js
"use strict"

module.exports = (callback, context) => {
    callback(null, {"message": "You said: " + context})
}
```

### Other languages...

[Sample functions](https://github.com/openfaas/faas/tree/master/sample-functions) in a range of other languages are available in the Github repository.

If there is a language then you would like to see added please contact the project team.