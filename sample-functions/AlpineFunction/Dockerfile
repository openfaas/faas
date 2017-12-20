FROM alpine:3.6

ADD https://github.com/openfaas/faas/releases/download/0.6.9/fwatchdog /usr/bin
# COPY ./fwatchdog /usr/bin/
RUN chmod +x /usr/bin/fwatchdog

# Populate example here
# ENV fprocess="wc -l"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["fwatchdog"]
