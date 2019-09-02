FROM golang:1.11-alpine as builder
ENV CGO_ENABLED=0

MAINTAINER alex@openfaas.com
ENTRYPOINT []

WORKDIR /go/src/github.com/openfaas/faas/sample-functions/WebhookStash

COPY handler.go .
# COPY vendor vendor

RUN go install

FROM functions/alpine:latest

# Needed to reach the hub
RUN apk --no-cache add ca-certificates 

COPY --from=builder /go/bin/WebhookStash  /usr/bin/WebhookStash
ENV fprocess "/usr/bin/WebhookStash"

CMD ["/usr/bin/fwatchdog"]
