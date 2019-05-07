FROM golang:1.10.4-alpine3.8 as builder

# Allows you to add additional packages via build-arg
ARG ADDITIONAL_PACKAGE
ARG CGO_ENABLED=0

RUN apk --no-cache add curl ${ADDITIONAL_PACKAGE} \
  && echo "Pulling watchdog binary from Github." \
  && curl -sSL https://github.com/openfaas/faas/releases/download/0.13.0/fwatchdog > /usr/bin/fwatchdog \
  && chmod +x /usr/bin/fwatchdog \
  && apk del curl --no-cache

WORKDIR /go/src/handler
COPY . .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./function/vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=linux \
  go build --ldflags "-s -w" -a -installsuffix cgo -o handler . && \
  go test $(go list ./... | grep -v /vendor/) -cover

FROM alpine:3.8
RUN apk --no-cache add ca-certificates

# Add non root user
RUN addgroup -S app && adduser -S -g app app
RUN mkdir -p /home/app

WORKDIR /home/app

COPY --from=builder /usr/bin/fwatchdog         .

COPY --from=builder /go/src/handler/function/  .
COPY --from=builder /go/src/handler/handler    .

ENV fprocess "/usr/bin/ApiKeyProtected"

RUN chown -R app /home/app

USER app

ENV fprocess="./handler"
EXPOSE 8080

HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1

CMD ["./fwatchdog"]
