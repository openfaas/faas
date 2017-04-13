FROM alpine:latest

WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

ADD https://github.com/alexellis/faas/releases/download/v0.5-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog
# COPY fwatchdog  /usr/bin/

COPY app    .

ENV fprocess="/root/app"
CMD ["fwatchdog"]
