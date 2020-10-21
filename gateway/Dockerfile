FROM teamserverless/license-check:0.3.6 as license-check

FROM golang:1.15-alpine as build

ENV GO11MODULE=off
ENV CGO_ENABLED=0

ARG GIT_COMMIT_SHA
ARG GIT_COMMIT_MESSAGE
ARG VERSION='dev'
ARG TARGETARCH

WORKDIR /go/src/github.com/openfaas/faas/gateway

COPY go.mod         .
COPY go.sum         .

COPY --from=license-check /license-check /usr/bin/

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
COPY main.go        .

RUN license-check -path ./ --verbose=false "Alex Ellis" "OpenFaaS Authors" "OpenFaaS Author(s)"

RUN test -z "$(gofmt -l $(find . -type f -name '*.go'))"

COPY vendor         vendor

RUN go test $(go list ./... | grep -v integration | grep -v /vendor/ | grep -v /template/) -cover

RUN go build -trimpath -ldflags="-buildid= -w -s \
        -X github.com/openfaas/faas/gateway/version.GitCommitSHA=${GIT_COMMIT_SHA}\
        -X github.com/openfaas/faas/gateway/version.GitCommitMessage=${GIT_COMMIT_MESSAGE}\
        -X github.com/openfaas/faas/gateway/version.Version=${VERSION} \
        -X github.com/openfaas/faas/gateway/types.Arch=${TARGETARCH}"\
    -a -installsuffix cgo -o gateway .

FROM alpine:3.12

ARG TARGETARCH

LABEL org.label-schema.license="MIT"
LABEL org.label-schema.vcs-url="https://github.com/openfaas/faas"
LABEL org.label-schema.vcs-type="Git"
LABEL org.label-schema.name="openfaas/faas"
LABEL org.label-schema.vendor="openfaas"
LABEL org.label-schema.docker.schema-version="1.0"

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

RUN echo ${TARGETARCH}
RUN sed -ie s/x86_64/${TARGETARCH}/g assets/script/funcstore.js \
    && rm assets/script/funcstore.jse

RUN chown -R app:app ./

USER app

CMD ["./gateway"]
