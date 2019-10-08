FROM openfaas/classic-watchdog:0.14.4 as watchdog

FROM alpine:3.10

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

RUN apk add --no-cache nmap

ENV fprocess="xargs nmap"


CMD ["fwatchdog"]
