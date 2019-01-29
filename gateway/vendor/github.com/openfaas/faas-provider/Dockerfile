FROM golang:1.10.4-alpine3.8

RUN mkdir -p /go/src/github.com/openfaas/faas-provider/

WORKDIR /go/src/github.com/openfaas/faas-provider

COPY vendor     vendor
COPY types      types
COPY auth       auth
COPY serve.go   .

RUN go test ./auth/ -v \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faas-provider .

FROM alpine:3.8
RUN apk --no-cache add ca-certificates
WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=0 /go/src/github.com/openfaas/faas-provider/faas-provider    .

CMD ["./faas-provider]
