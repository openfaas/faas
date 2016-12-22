FROM alpine:edge

COPY ./fwatchdog /usr/bin/fwatchdog
ENV fprocess "cat -b"

EXPOSE 8080
CMD ["fwatchdog"]
