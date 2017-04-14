FROM v4tech/imagemagick

ADD https://github.com/alexellis/faas/releases/download/0.5.2-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess "convert - -resize 50% fd:1"

EXPOSE 8080

CMD [ "/usr/bin/fwatchdog"]
