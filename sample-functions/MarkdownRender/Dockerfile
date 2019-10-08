FROM openfaas/classic-watchdog:0.14.4 as watchdog

FROM golang:1.11-alpine as builder
ENV CGO_ENABLED=0

MAINTAINER alex@openfaas.com
ENTRYPOINT []

WORKDIR /go/src/github.com/openfaas/faas/sample-functions/MarkdownRender

COPY handler.go .
COPY vendor vendor

RUN go install

FROM alpine:3.10

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

COPY --from=builder /go/bin/MarkdownRender  /usr/bin/MarkdownRender
ENV fprocess "/usr/bin/MarkdownRender"

CMD ["/usr/bin/fwatchdog"]
