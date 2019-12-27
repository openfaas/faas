FROM openfaas/classic-watchdog:0.18.8 as watchdog

FROM artemklevtsov/r-alpine:latest

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /application/

COPY handler.R .

ENV fprocess="Rscript handler.R"

RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app
USER 1000

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
