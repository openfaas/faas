FROM functions/alpine:latest

RUN apk add ffmpeg gifsicle
WORKDIR /root/
COPY entry.sh   .
ENV fprocess="./entry.sh"
CMD ["fwatchdog"]
