FROM golang:1.9.7-alpine as golang
WORKDIR /go/src/github.com/openfaas/nats-queue-worker

COPY vendor     vendor
COPY handler    handler
COPY main.go  .

RUN go test -v ./handler/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:3.7

RUN addgroup -S app \
  && adduser -S -g app app \
  && apk add --no-cache ca-certificates

WORKDIR /home/app

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=golang /go/src/github.com/openfaas/nats-queue-worker/app    .

RUN chown -R app:app ./

USER app
CMD ["./app"]
