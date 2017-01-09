FROM golang:1.7.3
FROM alpine:latest

COPY gateway    .

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

CMD ["./gateway"]
