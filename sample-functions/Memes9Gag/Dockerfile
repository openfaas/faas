FROM golang:1.9.1 AS builder
MAINTAINER mlabouardy <mohamed@labouardy.com>
WORKDIR /go/src/github.com/mlabouardy/Memes9Gag
RUN go get -d -v github.com/mlabouardy/9gag
COPY handler.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o handler .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/mlabouardy/Memes9Gag/handler .
CMD ["./handler"]
