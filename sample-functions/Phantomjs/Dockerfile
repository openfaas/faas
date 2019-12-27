FROM openfaas/classic-watchdog:0.18.8 as watchdog

FROM alexellis2/phantomjs-docker:latest

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess="phantomjs /dev/stdin"

RUN adduser --uid 1000 --group app
USER 1000

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["fwatchdog"]
