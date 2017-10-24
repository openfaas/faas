FROM golang:1.9.1 AS builder
MAINTAINER mlabouardy <mohamed@labouardy.com>
WORKDIR /go/src/github.com/mlabouardy/Memes9Gag
RUN go get -d -v github.com/mlabouardy/9gag
COPY handler.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
ADD https://github.com/openfaas/faas/releases/download/0.5.1-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog
WORKDIR /root/
COPY --from=builder /go/src/github.com/mlabouardy/Memes9Gag/app .
ENV fprocess="/root/app"
CMD ["fwatchdog"]
