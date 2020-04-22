FROM teamserverless/license-check:0.3.6 as license-check

FROM golang:1.13 as build

ENV GO111MODULE=off
ENV CGO_ENABLED=0

ARG GIT_COMMIT_SHA
ARG GIT_COMMIT_MESSAGE
ARG VERSION='dev'
ARG ARCH="x86_64"
ARG GOOS=linux
ARG GOARM

COPY --from=license-check /license-check /usr/bin/

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
COPY pkg            pkg
COPY main.go      .

# Run a gofmt and exclude all vendored code.
RUN license-check -path ./ --verbose=false "Alex Ellis" "OpenFaaS Authors" "OpenFaaS Author(s)" \
    && test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" \
    && CGO_ENABLED=${CGO_ENABLED} go test $(go list ./... | grep -v integration | grep -v /vendor/ | grep -v /template/) -cover \
    && GOARM=${GOARM} CGO_ENABLED=${CGO_ENABLED} GOOS=linux go build --ldflags "-s -w \
    -X github.com/openfaas/faas/gateway/version.GitCommitSHA=${GIT_COMMIT_SHA}\
    -X \"github.com/openfaas/faas/gateway/version.GitCommitMessage=${GIT_COMMIT_MESSAGE}\"\
    -X \"github.com/openfaas/faas/gateway/version.Version=${VERSION}\" \
    -X github.com/openfaas/faas/gateway/types.Arch=${ARCH}"\
    -a -installsuffix cgo -o gateway .

FROM alpine:3.11

ARG ARCH="x86_64"

LABEL org.label-schema.license="MIT" \
    org.label-schema.vcs-url="https://github.com/openfaas/faas" \
    org.label-schema.vcs-type="Git" \
    org.label-schema.name="openfaas/faas" \
    org.label-schema.vendor="openfaas" \
    org.label-schema.docker.schema-version="1.0"

RUN addgroup -S app \
    && adduser -S -g app app \
    && apk add --no-cache ca-certificates

WORKDIR /home/app

EXPOSE 8080
EXPOSE 8082
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /go/src/github.com/openfaas/faas/gateway/gateway    .
COPY assets     assets
RUN sed -ie s/x86_64/${ARCH}/g assets/script/funcstore.js && \
  rm assets/script/funcstore.jse

RUN chown -R app:app ./

USER app

CMD ["./gateway"]

