## Develop your own function

### TestDrive

Before you start development, you may want to take FaaS for a test drive which sets up a stack of sample functions from docker-compose.yml. You can then build your own functions and add them to the stack.

> You can test-drive FaaS with a set of sample functions as defined in docker-compose.yml on play-with-docker.com for free, or on your own laptop.

* [Begin the TestDrive instructions](https://github.com/openfaas/faas/blob/master/TestDrive.md)

### Working on the API Gateway or Watchdog

To work on either of the FaaS Golang components checkout the "./build.sh" scripts and acompanying Dockerfiles.

* [Roadmap and Contributing](https://github.com/openfaas/faas/blob/master/ROADMAP.md)

### Creating a function

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
        networks:
            - functions
        environment:
            fprocess:	"wc"
```

### Testing your function

You can test your function through a webbrowser against the UI portal on port 8080.

http://localhost:8080/

You can also invoke a function by name with curl:

```
curl --data-binary @README.md http://localhost:8080/function/func_wordcount
```

