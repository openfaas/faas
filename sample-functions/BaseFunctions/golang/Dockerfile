FROM golang:1.9.4-alpine

MAINTAINER alexellis2@gmail.com
ENTRYPOINT []

WORKDIR /go/src/github.com/openfaas/faas/sample-functions/golang
COPY . /go/src/github.com/openfaas/faas/sample-functions/golang

RUN go install

ADD https://github.com/openfaas/faas/releases/download/0.7.1/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess "/go/bin/golang"
HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD [ "/usr/bin/fwatchdog"]
