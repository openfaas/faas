FROM --platform=${BUILDPLATFORM:-linux/amd64} teamserverless/license-check:0.3.9 as license-check

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.13 as build

ENV GO111MODULE=off
ENV CGO_ENABLED=0

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG GIT_COMMIT
ARG GIT_COMMIT_MESSAGE
ARG VERSION

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
COPY main.go        .
RUN license-check -path ./ --verbose=false "Alex Ellis" "OpenFaaS Authors" "OpenFaaS Author(s)"


# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))"
RUN go test $(go list ./... | grep -v integration | grep -v /vendor/ | grep -v /template/) -cover

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build --ldflags "-s -w \
    -X \"github.com/openfaas/faas/gateway/version.GitCommitSHA=${GIT_COMMIT}\" \
    -X \"github.com/openfaas/faas/gateway/version.GitCommitMessage=${GIT_COMMIT_MESSAGE}\" \
    -X \"github.com/openfaas/faas/gateway/version.Version=${VERSION}\" \
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
COPY assets     assets

ARG TARGETPLATFORM

RUN if [ "$TARGETPLATFORM" = "linux/arm/v7" ] ; then sed -ie s/x86_64/armhf/g assets/script/funcstore.js ; elif [ "$TARGETPLATFORM" = "linux/arm64" ] ; then sed -ie s/x86_64/arm64/g assets/script/funcstore.js; fi

RUN chown -R app:app ./

USER app

CMD ["./gateway"]

