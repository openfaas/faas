FROM --platform=${TARGETPLATFORM:-linux/amd64} ghcr.io/openfaas/classic-watchdog:0.1.4 as watchdog

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.15.4

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

# Populate example here
# ENV fprocess="wc -l"

RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app
USER 1000

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["fwatchdog"]
