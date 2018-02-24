FROM v4tech/imagemagick

ADD https://github.com/openfaas/faas/releases/download/0.7.1/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess "convert - -resize 50% fd:1"

EXPOSE 8080

CMD [ "/usr/bin/fwatchdog"]
