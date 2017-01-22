FROM alpine:latest

WORKDIR /root/

COPY gateway    .
COPY assets assets

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

CMD ["./gateway"]
