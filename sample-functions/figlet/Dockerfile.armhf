FROM functions/alpine:latest

RUN apk add --no-cache figlet

ENV fprocess="figlet"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["fwatchdog"]