FROM openfaas/classic-watchdog:0.14.4 as watchdog

FROM golang:1.11-alpine as builder
ENV CGO_ENABLED=0

MAINTAINER alex@openfaas.com
ENTRYPOINT []

WORKDIR /go/src/github.com/openfaas/faas/sample-functions/WebhookStash

COPY handler.go .

RUN go install

FROM alpine:3.10

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

COPY --from=builder /go/bin/WebhookStash  /usr/bin/WebhookStash
ENV fprocess "/usr/bin/WebhookStash"

CMD ["/usr/bin/fwatchdog"]
