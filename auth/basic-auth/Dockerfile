FROM --platform=${BUILDPLATFORM:-linux/amd64} teamserverless/license-check:0.3.6 as license-check

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.13 as build

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
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH } test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))"
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH } \
    go build --ldflags "-s -w" -a -installsuffix cgo -o handler . && \
    go test $(go list ./... | grep -v /vendor/) -cover

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.12 as ship
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
