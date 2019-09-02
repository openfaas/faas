FROM golang:1.11-alpine as builder
ENV CGO_ENABLED=0

MAINTAINER alex@openfaas.com
ENTRYPOINT []

WORKDIR /go/src/github.com/openfaas/faas/sample-functions/MarkdownRender

COPY handler.go .
COPY vendor vendor

RUN go install

FROM functions/alpine:latest

# Needed to reach the hub
RUN apk --no-cache add ca-certificates 

COPY --from=builder /go/bin/MarkdownRender  /usr/bin/MarkdownRender
ENV fprocess "/usr/bin/MarkdownRender"

CMD ["/usr/bin/fwatchdog"]
