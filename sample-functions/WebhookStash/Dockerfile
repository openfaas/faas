FROM golang:1.9.7-alpine as builder

MAINTAINER alex@openfaas.com
ENTRYPOINT []

RUN apk --no-cache add make curl \
    && curl -sL https://github.com/openfaas/faas/releases/download/0.9.6/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

WORKDIR /go/src/github.com/openfaas/faas/sample-functions/WebhookStash

COPY handler.go .
# COPY vendor vendor

RUN go install

FROM alpine:3.7

# Needed to reach the hub
RUN apk --no-cache add ca-certificates 

COPY --from=builder /usr/bin/fwatchdog  /usr/bin/fwatchdog
COPY --from=builder /go/bin/WebhookStash  /usr/bin/WebhookStash
ENV fprocess "/usr/bin/WebhookStash"

CMD ["/usr/bin/fwatchdog"]
