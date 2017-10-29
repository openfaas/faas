FROM debian:stretch

RUN mkdir -p /usr/local/go
ENV PATH=$PATH:/usr/local/go/bin
RUN apt update && apt -qy install curl \
 && curl -SL https://storage.googleapis.com/golang/go1.9.linux-arm64.tar.gz > go1.9.linux-arm64.tar.gz \
 && tar -xvf go1.9.linux-arm64.tar.gz  -C /usr/local/go --strip-components=1

RUN go version

CMD ["/bin/sh"]
