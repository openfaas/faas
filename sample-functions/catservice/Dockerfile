FROM alpine:latest
ENTRYPOINT []

ADD https://github.com/alexellis/faas/releases/download/v0.5-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess "/bin/cat"

EXPOSE 8080
CMD ["/usr/bin/fwatchdog"]
