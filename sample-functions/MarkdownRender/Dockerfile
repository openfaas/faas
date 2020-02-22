FROM openfaas/classic-watchdog:0.18.8 as watchdog

FROM golang:1.13-alpine as builder
ENV CGO_ENABLED=0

MAINTAINER alex@openfaas.com
ENTRYPOINT []

WORKDIR /go/src/github.com/openfaas/faas/sample-functions/MarkdownRender

COPY handler.go .
COPY vendor vendor

RUN go install

FROM alpine:3.11

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

COPY --from=builder /go/bin/MarkdownRender  /usr/bin/MarkdownRender
ENV fprocess "/usr/bin/MarkdownRender"

RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app
USER 1000

CMD ["/usr/bin/fwatchdog"]
