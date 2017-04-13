FROM alpine:latest
ENTRYPOINT []

ADD https://github.com/alexellis/faas/releases/download/0.5.3-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess "/bin/cat"

EXPOSE 8080
CMD ["/usr/bin/fwatchdog"]
