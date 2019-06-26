FROM openfaas/classic-watchdog:0.14.4 as watchdog

FROM v4tech/imagemagick

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess "convert - -resize 50% fd:1"

RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app
USER 1000

EXPOSE 8080

CMD [ "/usr/bin/fwatchdog"]
