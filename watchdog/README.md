Watchdog
==========

The watchdog provides an unmanaged and generic interface between the outside world and your function. Its job is to marshal a HTTP request accepted on the API Gateway and to invoke your chosen application. The watchdog is a tiny Golang webserver - see the diagram below for how this process works.

![](https://pbs.twimg.com/media/DGScDblUIAAo4H-.jpg:large)

>  Above: a tiny web-server or shim that forks your desired process for every incoming HTTP request

Every function needs to embed this binary and use it as its `ENTRYPOINT` or `CMD`, in effect it is the init process for your container. Once your process is forked the watchdog passses in the HTTP request via `stdin` and reads a HTTP response via `stdout`. This means your process does not need to know anything about the web or HTTP.

### Next-gen: of-watchdog

Are you looking for more control over your HTTP responses, "hot functions", persistent connection pools or to cache a machine-learning model in memory? Then check out the *http mode* of the new [of-watchdog](https://github.com/openfaas-incubator/of-watchdog).

## Create a new function the easy way

**Create a function via the CLI**

The easiest way to create a function is to use a template and the FaaS CLI. The CLI allows you to abstract all Docker knowledge away, you just have to write a handler file in one of the supported programming languages.

* [Your first serverless Python function with OpenFaaS](https://blog.alexellis.io/first-faas-python-function/)

* [Read a tutorial on the FaaS CLI](https://github.com/openfaas/faas-cli)

## Delve deeper

**Package your function**

Here's how to package your function if you don't want to use the CLI or have existing binaries or images:

- [x] Use an existing or a new Docker image as base image `FROM`
- [x] Add the fwatchdog binary from the [Releases page](https://github.com/openfaas/faas/releases) via `curl` or `ADD https://`
- [x] Set an `fprocess` (function process) environmental variable with the function you want to run for each request
- [x] Expose port 8080
- [x] Set the `CMD` to `fwatchdog`

Example Dockerfile for an `echo` function:

```
FROM alpine:3.8

ADD https://github.com/openfaas/faas/releases/download/0.9.6/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

# Define your binary here
ENV fprocess="/bin/cat"

CMD ["fwatchdog"]
```

**Tip:**
You can optimize Docker to cache getting the watchdog by using curl, instead of ADD.
To do so, replace the related lines with:
```
RUN apk --no-cache add curl \
    && curl -sL https://github.com/openfaas/faas/releases/download/0.9.6/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog
```

**Implementing a health-check**

At any point in time, if you detect that your function has become unhealthy and needs to restart, then you can delete the `/tmp/.lock` file which invalidates the check and causes Swarm to re-schedule the function.

* Kubernetes

For Kubernetes the health check is added through automation without you needing to alter the `Dockerfile`.

* Swarm

A Docker Swarm Healthcheck is required and is best practice. It will make sure that the watchdog is ready to accept a request before forwarding requests via the API Gateway. If the function or watchdog runs into an unrecoverable issue Swarm will also be able to restart the container.

Here is an example of the `echo` function implementing a healthcheck with a 5-second checking interval.

```
FROM functions/alpine

ENV fprocess="cat /etc/hostname"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
```

The watchdog process creates a .lock file in `/tmp/` on starting its internal Golang HTTP server. `[ -e file_name ]` is shell to check if a file exists. With Windows Containers this is an invalid path so you may want to set the `suppress_lock` environmental variable.

Read my Docker Swarm tutorial on Healthchecks:

 * [Test-drive Docker Healthcheck in 10 minutes](http://blog.alexellis.io/test-drive-healthcheck/)

**Environmental overrides:**

The watchdog can be configured through environmental variables. You must always specifiy an `fprocess` variable.

| Option                 | Usage             |
|------------------------|--------------|
| `fprocess`             | The process to invoke for each function call (function process). This must be a UNIX binary and accept input via STDIN and output via STDOUT  |
| `cgi_headers`          | HTTP headers from request are made available through environmental variables - `Http_X_Served_By` etc. See section: *Handling headers* for more detail. Enabled by default |
| `marshal_request`     | Instead of re-directing the raw HTTP body into your fprocess, it will first be marshalled into JSON. Use this if you need to work with HTTP headers and do not want to use environmental variables via the `cgi_headers` flag. |
| `content_type`         | Force a specific Content-Type response for all responses |
| `write_timeout`        | HTTP timeout for writing a response body from your function (in seconds)  |
| `read_timeout`         | HTTP timeout for reading the payload from the client caller (in seconds) |
| `suppress_lock`        | The watchdog will attempt to write a lockfile to /tmp/ for swarm healthchecks - set this to true to disable behaviour. |
| `exec_timeout`         | Hard timeout for process exec'd for each incoming request (in seconds). Disabled if set to 0 |
| `write_debug`          | Write all output, error messages, and additional information to the logs. Default is false |
| `combine_output`       | True by default - combines stdout/stderr in function response, when set to false `stderr` is written to the container logs and stdout is used for function response |

## Advanced / tuning

### (New) of-watchdog and HTTP mode

* of-watchdog

Forking a new process per request has advantages such as process isolation, portability and simplicity. Any process can be made into a function without any additional code. The of-watchdog and its "HTTP" mode is an optimization which maintains one single process between all requests.

A new version of the watchdog is being tested over at [openfaas-incubator/of-watchdog](https://github.com/openfaas-incubator/of-watchdog).

This re-write is mainly structural for on-going maintenance. It will be a drop-in replacement for the existing watchdog and also has binary releases available.

### Graceful shutdowns

The watchdog is capable of working with health-checks to provide a graceful shutdown.

When a `SIGTERM` signal is detected within the watchdog process a Go routine will remove the `/tmp/.lock` file and mark the HTTP health-check as unhealthy and return HTTP 503. The code will then wait for the duration specified in `write_timeout`. During this window the container-orchestrator's health-check must run and complete.

Now the orchestrator will mark this replica as unhealthy and remove it from the pool of valid HTTP endpoints.

Now we will stop accepting new connections and wait for the value defined in `write_timeout` before finally allowing the process to exit.

### Working with HTTP headers

Headers and other request information are injected into environmental variables in the following format:

The `X-Forwarded-By` header becomes available as `Http_X_Forwarded_By`

* `Http_Method` - GET/POST etc
* `Http_Query` - QueryString value
* `Http_ContentLength` - gives the total content-length of the incoming HTTP request received by the watchdog.

> This behaviour is enabled by the `cgi_headers` environmental variable which is enabled by default.

Here's an example of a POST request with an additional header and a query-string.

```
$ cgi_headers=true fprocess=env ./watchdog &
2017/06/23 17:02:58 Writing lock-file to: /tmp/.lock

$ curl "localhost:8080?q=serverless&page=1" -X POST -H X-Forwarded-By:http://my.vpn.com
```

This is what you'd see if you had set your `fprocess` to `env` on a Linux system:

```
Http_User_Agent=curl/7.43.0
Http_Accept=*/*
Http_X_Forwarded_By=http://my.vpn.com
Http_Method=POST
Http_Query=q=serverless&page=1
```

You can also use the `GET` verb:

```
$ curl "localhost:8080?action=quote&qty=1&productId=105"
```

The output from the watchdog would be:

```
Http_User_Agent=curl/7.43.0
Http_Accept=*/*
Http_Method=GET
Http_Query=action=quote&qty=1&productId=105
```

You can now use HTTP state from within your application to make decisions.

### HTTP methods

The HTTP methods supported for the watchdog are:

With a body:
* POST, PUT, DELETE, UPDATE

Without a body:
* GET

> The API Gateway currently supports the POST route for functions.

### Content-Type of request/response

By default the watchdog will match the response of your function to the "Content-Type" of the client.

* If your client sends a JSON post with a Content-Type of `application/json` this will be matched automatically in the response.
* If your client sends a JSON post with a Content-Type of `text/plain` this will be matched automatically in the response too

To override the Content-Type of all your responses set the `content_type` environmental variable.

### I don't want to use the watchdog

This is an unsupported use-case for the OpenFaaS project however if your container conforms to the requirements below then the OpenFaaS API gateway and other tooling will manage and scale your service.

You will need to provide a lock-file at `/tmp/.lock` so that the orchestration system can run healthchecks on your container. If you are using Docker Swarm make sure you provide a `HEALTHCHECK` instruction in your Dockerfile - samples are given in the `faas` repository.

* Expose TCP port 8080 over HTTP
* Create `/tmp/.lock` or in whatever location responds to the OS tempdir syscall

### Tuning auto-scaling

Auto-scaling starts at 1 replica and steps up in blocks of 5:

* 1->5
* 5->10
* 10->15
* 15->20

You can override the minimum and maximum scale of a function through labels.

Add these labels to the deployment if you want to sacle between 2 and 15 replicas.

```
com.openfaas.scale.min: "2"
com.openfaas.scale.max: "15"
```

The labels are optional.

**Disabling auto-scaling**

If you want to disable auto-scaling for a function then set the minimum and maximum scale to the same value i.e. "1".

As an alternative you can also remove AlertManager or scale it to 0 replicas.
