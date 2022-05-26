FROM ghcr.io/openfaas/classic-watchdog:0.2.0 as watchdog

FROM alpine:3.16.0 as ship

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

RUN apk add --no-cache nmap

ENV fprocess="xargs nmap"

RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app
USER 1000

CMD ["fwatchdog"]
