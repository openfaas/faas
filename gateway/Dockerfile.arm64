FROM alexellis2/golang:1.9-arm64 as build
WORKDIR /go/src/github.com/openfaas/faas/gateway
ENV GOPATH=/go

COPY vendor         vendor

COPY handlers       handlers
COPY metrics        metrics
COPY requests       requests
COPY tests          tests

COPY types          types
COPY queue          queue
COPY plugin         plugin
COPY server.go      .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway .

FROM debian:stretch
WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /go/src/github.com/openfaas/faas/gateway/gateway    .

COPY assets     assets

CMD ["./gateway"]
