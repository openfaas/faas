FROM golang:1.7.5 as build
WORKDIR /go/src/github.com/alexellis/faas/gateway

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

FROM alpine:latest
WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /go/src/github.com/alexellis/faas/gateway/gateway    .

COPY assets     assets

CMD ["./gateway"]
