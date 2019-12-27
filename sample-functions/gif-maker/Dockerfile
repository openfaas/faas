FROM functions/alpine:latest

USER root

RUN apk add ffmpeg gifsicle
WORKDIR /root/
COPY entry.sh   .
ENV fprocess="./entry.sh"

USER 1000

CMD ["fwatchdog"]
