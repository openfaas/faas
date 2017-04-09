FROM golang:1.7.5-alpine

MAINTAINER alexellis2@gmail.com
ENTRYPOINT []

WORKDIR /go/src/github.com/alexellis/faas/sample-functions/golang
COPY . /go/src/github.com/alexellis/faas/sample-functions/golang

RUN go install

ADD https://github.com/alexellis/faas/releases/download/v0.5-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess "/go/bin/golang"
HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD [ "/usr/bin/fwatchdog"]
