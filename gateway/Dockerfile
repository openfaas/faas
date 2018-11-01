FROM golang:1.9.7 as build
ARG GIT_COMMIT_SHA
ARG GIT_COMMIT_MESSAGE
ARG VERSION='dev'

RUN curl -sSfL https://github.com/alexellis/license-check/releases/download/0.2.2/license-check \
      > /usr/bin/license-check \
      && chmod +x /usr/bin/license-check

WORKDIR /go/src/github.com/openfaas/faas/gateway

COPY vendor         vendor

COPY handlers       handlers
COPY metrics        metrics
COPY requests       requests
COPY tests          tests

COPY types          types
COPY queue          queue
COPY plugin         plugin
COPY version        version
COPY scaling        scaling
COPY server.go      .

# Run a gofmt and exclude all vendored code.
RUN license-check -path ./ --verbose=false "Alex Ellis" "OpenFaaS Project" "OpenFaaS Authors" "OpenFaaS Author(s)" \
    && test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" \
    && go test $(go list ./... | grep -v integration | grep -v /vendor/ | grep -v /template/) -cover \
    && CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w \
     -X github.com/openfaas/faas/gateway/version.GitCommitSHA=${GIT_COMMIT_SHA}\
     -X \"github.com/openfaas/faas/gateway/version.GitCommitMessage=${GIT_COMMIT_MESSAGE}\"\
     -X github.com/openfaas/faas/gateway/version.Version=${VERSION}" \
    -a -installsuffix cgo -o gateway .

FROM alpine:3.7

LABEL org.label-schema.license="MIT" \
      org.label-schema.vcs-url="https://github.com/openfaas/faas" \
      org.label-schema.vcs-type="Git" \
      org.label-schema.name="openfaas/faas" \
      org.label-schema.vendor="openfaas" \
      org.label-schema.docker.schema-version="1.0"

RUN addgroup -S app \
    && adduser -S -g app app

WORKDIR /home/app

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /go/src/github.com/openfaas/faas/gateway/gateway    .
COPY assets     assets

RUN chown -R app:app ./

USER app

CMD ["./gateway"]
