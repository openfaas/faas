FROM golang:1.11-alpine3.10 as build
ENV CGO_ENABLED=0

RUN apk add --no-cache curl ca-certificates
RUN curl -sLSf https://raw.githubusercontent.com/teamserverless/license-check/master/get.sh | sh \
  && mv ./license-check /usr/bin/

WORKDIR /go/src/handler
COPY . .

# Run a gofmt and exclude all vendored code.

RUN license-check -path ./ --verbose=false "OpenFaaS Authors" "OpenFaaS Author(s)" \
 && test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" \
 && CGO_ENABLED=0 GOOS=linux \
    go build --ldflags "-s -w" -a -installsuffix cgo -o handler . && \
    go test $(go list ./... | grep -v /vendor/) -cover

FROM alpine:3.10 as ship
# Add non-root user
RUN addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

COPY --from=build /go/src/handler/handler    .

RUN chown -R app /home/app

USER app

WORKDIR /home/app

CMD ["./handler"]
