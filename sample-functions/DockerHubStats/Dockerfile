FROM openfaas/classic-watchdog:0.14.4 as watchdog

FROM golang:1.9.7-alpine as builder

MAINTAINER alex@openfaas.com
ENTRYPOINT []

WORKDIR /go/src/github.com/openfaas/faas/sample-functions/DockerHubStats
COPY . /go/src/github.com/openfaas/faas/sample-functions/DockerHubStats
RUN set -ex && apk add make && make install

FROM alpine:3.10

# Needed to reach the hub
RUN apk --no-cache add ca-certificates 

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog
COPY --from=builder /go/bin/DockerHubStats  /usr/bin/DockerHubStats
ENV fprocess "/usr/bin/DockerHubStats"

RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app
USER 1000

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["/usr/bin/fwatchdog"]
