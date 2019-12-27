FROM functions/alpine:latest
USER root

RUN apk add --no-cache figlet

USER 1000
ENV fprocess="figlet"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["fwatchdog"]