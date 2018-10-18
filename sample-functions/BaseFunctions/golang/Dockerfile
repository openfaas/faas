FROM golang:1.9.7-alpine

MAINTAINER alexellis2@gmail.com
ENTRYPOINT []

WORKDIR /go/src/github.com/openfaas/faas/sample-functions/golang
COPY . /go/src/github.com/openfaas/faas/sample-functions/golang

RUN go install

RUN apk --no-cache add curl \
    && curl -sL https://github.com/openfaas/faas/releases/download/0.9.6/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

ENV fprocess "/go/bin/golang"
HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD [ "/usr/bin/fwatchdog"]
