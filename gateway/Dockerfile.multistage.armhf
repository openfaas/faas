FROM alexellis2/go-armhf:1.7.4

WORKDIR /go/src/github.com/alexellis/faas/gateway

COPY vendor         vendor

COPY handlers       handlers
COPY metrics        metrics
COPY requests       requests
COPY tests          tests
COPY server.go      .
COPY types          types
COPY plugin  plugin

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway .

FROM armhf/alpine:latest
WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=0 /go/src/github.com/alexellis/faas/gateway/gateway    .

COPY assets     assets

CMD ["./gateway"]

