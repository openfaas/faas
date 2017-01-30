FROM aafrey/golang-alpine:arm
ENTRYPOINT []

COPY ./fwatchdog /usr/bin/fwatchdog
ENV fprocess "cat"


EXPOSE 8080
CMD ["/usr/bin/fwatchdog"]
