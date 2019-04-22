FROM v4tech/imagemagick

RUN apk --no-cache add curl \
    && curl -sL https://github.com/openfaas/faas/releases/download/0.13.0/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

ENV fprocess "convert - -resize 50% fd:1"

EXPOSE 8080

CMD [ "/usr/bin/fwatchdog"]
