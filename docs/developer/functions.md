# Functions

## Creating a function

Functions run as Docker containers with the Watchdog component embedded to handle communication with the API Gateway.

You can find the [reference documentation for the Watchdog here](https://github.com/openfaas/faas/tree/master/watchdog).


**Markdown Parser**

This is the basis of a function which generates HTML from MarkDown:

```
FROM golang:1.7.5
RUN mkdir -p /go/src/app
COPY handler.go /go/src/app
WORKDIR /go/src/app
RUN go get github.com/microcosm-cc/bluemonday && \
    go get github.com/russross/blackfriday

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

ADD https://github.com/openfaas/faas/releases/download/v0.3-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess="/go/src/app/app"

CMD ["/usr/bin/fwatchdog"]
```

The base Docker container is not important, you just need to add the watchdog component and then set the fprocess to execute your binary at runtime.

Update the Docker stack with this:

```
    markdown:
        image: alexellis2/faas-markdownrender:latest
        labels:
            function: "true"
        depends_on:
            - gateway
        networks:
            - functions
```

**Word counter with busybox**

```
FROM alpine:latest

ADD https://github.com/openfaas/faas/releases/download/v0.3-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess="wc"
CMD ["fwatchdog"]
```

Update your Docker stack with this definition:

```
    wordcount:
        image: alexellis2/faas-alpinefunction:latest
        labels:
            function: "true"
        depends_on:
            - gateway
        networks:
            - functions
        environment:
            fprocess:	"wc"
```

## Testing your function

You can test your function through a webbrowser against the UI portal on port 8080.

http://localhost:8080/

You can also invoke a function by name with curl:

```
curl --data-binary @README.md http://localhost:8080/function/func_wordcount
```

## Asynchronous processing

By default functions are accessed synchronously via the following route:

```
$ curl --data "message" http://gateway/function/{function_name}
```

As of [PR #131](https://github.com/openfaas/faas/pull/131) asynchronous invocation is available for testing.

*Logical flow for synchronous functions:*

![](https://user-images.githubusercontent.com/6358735/29469107-cbc38c88-843e-11e7-9516-c0dd33bab63b.png)

### Why use Asynchronous processing?

* Enable longer time-outs

* Process work whenever resources are available rather than immediately

* Consume a large batch of work within a few seconds and let it process at its own pace

### How does async work?

Here is a conceptual diagram

![](https://user-images.githubusercontent.com/6358735/29469109-cc03c244-843e-11e7-9dfd-a540799dac28.png)

* [queue-worker](https://github.com/open-faas/nats-queue-worker)

### Deploy the async stack

The reference implementation for asychronous processing uses NATS Streaming, but you are free to extend OpenFaaS and write your own [queue-worker](https://github.com/open-faas/nats-queue-worker).

Swarm:

```
$ ./deploy_extended.sh
```

K8s:

```
$ kubectl -f delete ./faas.yml
$ kubectl -f apply ./faas.async.yml,nats.yml
```

### Call a function

Functions do not need to be modified to work asynchronously, just use this alternate route:

```
$ http://gateway/async-function/{function_name}
```

If you want the function to call another function or a different endpoint when it is finished then pass the `X-Callback-Url` header. This is optional.

```
$ curl http://gateway/async-function/{function_name} \
    -H "X-Callback-Url: http://gateway/function/send2slack" \
    --data-binary @sample.json
```

### Extend function timeouts

Functions have three timeouts configurable by environmental variables expressed in seconds:

HTTP:

* read_timeout
* write_timeout

Hard timeout:

* exec_timeout

To make use of these just add them to your Dockerfile when needed as ENV variables.

> [Function watchdog reference](https://github.com/openfaas/faas/tree/master/watchdog)


