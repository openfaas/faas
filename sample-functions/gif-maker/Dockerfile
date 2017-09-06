FROM alpine:3.6

RUN apk --no-cache add ffmpeg gifsicle curl \
   && curl -LS https://github.com/alexellis/faas/releases/download/0.5.6-alpha/fwatchdog > /usr/bin/fwatchdog \
   && chmod +x /usr/bin/fwatchdog \
   && apk del curl
WORKDIR /root/
COPY entry.sh   .
ENV fprocess="./entry.sh"
CMD ["fwatchdog"]
