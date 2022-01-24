FROM --platform=${BUILDPLATFORM:-linux/amd64} ghcr.io/openfaas/license-check:0.4.0 as license-check

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.17 as build

ENV GO111MODULE=off
ENV CGO_ENABLED=0

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY --from=license-check /license-check /usr/bin/

WORKDIR /go/src/handler
COPY . .

# Run a gofmt and exclude all vendored code.

RUN license-check -path ./ --verbose=false "OpenFaaS Authors" "OpenFaaS Author(s)"

RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))"

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go test -v ./...

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    --ldflags "-s -w" -a -installsuffix cgo -o handler .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.14 as ship
# Add non-root user
RUN addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app

COPY --from=build /go/src/handler/handler    .

RUN chown -R app /home/app

USER app

WORKDIR /home/app

CMD ["./handler"]
