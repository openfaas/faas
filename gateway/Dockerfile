FROM --platform=${BUILDPLATFORM:-linux/amd64} teamserverless/license-check:0.3.6 as license-check

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.13 as build

ENV GO111MODULE=off
ENV CGO_ENABLED=0

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG GIT_COMMIT_SHA
ARG GIT_COMMIT_MESSAGE
ARG VERSION='dev'

COPY --from=license-check /license-check /usr/bin/

WORKDIR /go/src/github.com/openfaas/faas/gateway

COPY gateway/vendor         vendor

COPY gateway/handlers       handlers
COPY gateway/metrics        metrics
COPY gateway/requests       requests
COPY gateway/tests          tests

COPY gateway/types          types
COPY gateway/queue          queue
COPY gateway/plugin         plugin
COPY gateway/version        version
COPY gateway/scaling        scaling
COPY gateway/pkg            pkg
COPY gateway/main.go        .
COPY .git .
RUN license-check -path ./ --verbose=false "Alex Ellis" "OpenFaaS Authors" "OpenFaaS Author(s)"


# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))"
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go test $(go list ./... | grep -v integration | grep -v /vendor/ | grep -v /template/) -cover
RUN GIT_COMMIT_MESSAGE=$(git log -1 --pretty=%B 2>&1 | head -n 1) \
    GIT_COMMIT_SHA=$(git rev-list -1 HEAD) \
    VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///' || echo dev) \
    CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build --ldflags "-s -w \
    -X github.com/openfaas/faas/gateway/version.GitCommitSHA=${GIT_COMMIT_SHA} \
    -X github.com/openfaas/faas/gateway/version.GitCommitMessage=${GIT_COMMIT_MESSAGE} \
    -X github.com/openfaas/faas/gateway/version.Version=${VERSION} \
    -X github.com/openfaas/faas/gateway/types.Arch=${TARGETARCH}" \
    -a -installsuffix cgo -o gateway .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.12 as ship

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
COPY gateway/assets     assets
RUN sed -ie s/x86_64/${GOARCH}/g assets/script/funcstore.js && \
  rm assets/script/funcstore.jse

RUN chown -R app:app ./

USER app

CMD ["./gateway"]

