FROM golang:1.8.5 as build
WORKDIR /go/src/github.com/openfaas/faas/gateway

RUN curl -sL https://github.com/alexellis/license-check/releases/download/0.1/license-check > /usr/bin/license-check \
    && chmod +x /usr/bin/license-check

COPY vendor         vendor

COPY handlers       handlers
COPY metrics        metrics
COPY requests       requests
COPY tests          tests

COPY types          types
COPY queue          queue
COPY plugin         plugin
COPY server.go      .

# Run a gofmt and exclude all vendored code.
RUN license-check -path ./ --verbose=false \
    && test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" \
    && go test $(go list ./... | grep -v -e integration -e /vendor/ -e /template/) -cover \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway .

FROM alpine:3.6
WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /go/src/github.com/openfaas/faas/gateway/gateway    .

COPY assets     assets

# Replace @@placeholder@@ in assets/index.html
RUN hash=$(sha1sum assets/style/bootstrap.css | awk '{print $1}') \
    && sed s/@@bootstrap.css@@/${hash}/ -i assets/index.html \
    && hash=$(sha1sum assets/script/bootstrap.js | awk '{print $1}') \
    && sed s/@@bootstrap.js@@/${hash}/ -i assets/index.html \
    && hash=$(sha1sum assets/script/funcstore.js | awk '{print $1}') \
    && sed s/@@funcstore.js@@/${hash}/ -i assets/index.html

CMD ["./gateway"]
