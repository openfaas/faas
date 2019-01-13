FROM alexellis2/phantomjs-docker:latest

RUN apt-get update && apt-get install -y curl \
    && curl -sL https://github.com/openfaas/faas/releases/download/0.9.14/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV fprocess="phantomjs /dev/stdin"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["fwatchdog"]
