FROM arm64v8/alpine:3.6

ADD https://github.com/alexellis/faas/releases/download/0.6.9/fwatchdog-arm64 /usr/bin/fwatchdog
# COPY ./fwatchdog /usr/bin/
RUN chmod +x /usr/bin/fwatchdog

# Populate example here
# ENV fprocess="wc -l"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["fwatchdog"]
