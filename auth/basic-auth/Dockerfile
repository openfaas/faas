FROM teamserverless/license-check:0.3.6 as license-check

FROM golang:1.13-alpine3.11 as build

ENV GO111MODULE=off
ENV CGO_ENABLED=0

RUN apk add --no-cache curl ca-certificates
COPY --from=license-check /license-check /usr/bin/

WORKDIR /go/src/handler
COPY . .

# Run a gofmt and exclude all vendored code.

RUN license-check -path ./ --verbose=false "OpenFaaS Authors" "OpenFaaS Author(s)" \
 && test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" \
 && CGO_ENABLED=0 GOOS=linux \
    go build --ldflags "-s -w" -a -installsuffix cgo -o handler . && \
    go test $(go list ./... | grep -v /vendor/) -cover

FROM alpine:3.11 as ship
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
