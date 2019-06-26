FROM openfaas/classic-watchdog:0.14.4 as watchdog

FROM alexellis2/phantomjs-docker:latest

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess="phantomjs /dev/stdin"

RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app
USER 1000

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["fwatchdog"]
