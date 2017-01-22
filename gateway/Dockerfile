FROM alpine:latest

WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY gateway    .
COPY assets     assets

CMD ["./gateway"]
