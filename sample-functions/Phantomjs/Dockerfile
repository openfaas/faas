FROM alexellis2/phantomjs-docker:latest

ADD https://github.com/openfaas/faas/releases/download/0.5.6-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess="phantomjs /dev/stdin"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["fwatchdog"]
