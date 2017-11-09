FROM golang:1.7.5 as builder
RUN mkdir -p /go/src/app
COPY handler.go /go/src/app
WORKDIR /go/src/app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .


FROM alpine:latest

WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

ADD https://github.com/alexellis/faas/releases/download/0.6.6d/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

COPY --from=builder /go/src/app/app    .

ENV fprocess="/root/app"
CMD ["fwatchdog"]
