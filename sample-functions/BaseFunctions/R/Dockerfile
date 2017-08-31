FROM artemklevtsov/r-alpine:latest

ADD https://github.com/alexellis/faas/releases/download/0.6.1/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /root/

COPY handler.R .

ENV fprocess="Rscript handler.R"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

