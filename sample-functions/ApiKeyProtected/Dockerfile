FROM alpine:latest

WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

ADD https://github.com/alexellis/faas/releases/download/0.5.1-alpha/fwatchdog /usr/bin
# COPY fwatchdog  /usr/bin/
RUN chmod +x /usr/bin/fwatchdog

COPY app    .

ENV fprocess="/root/app"
CMD ["fwatchdog"]
