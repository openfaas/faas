FROM openfaas/classic-watchdog:0.18.8 as watchdog

FROM golang:1.13-alpine
ENV CGO_ENABLED=0

MAINTAINER alexellis2@gmail.com
ENTRYPOINT []

WORKDIR /go/src/github.com/openfaas/faas/sample-functions/golang
COPY . /go/src/github.com/openfaas/faas/sample-functions/golang

RUN go install

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess "/go/bin/golang"
HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app
USER 1000

CMD [ "fwatchdog"]
