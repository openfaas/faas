FROM toricls/gnucobol:latest

ADD https://github.com/openfaas/faas/releases/download/0.5.1-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /root/

COPY handler.cob    .
RUN cobc -x handler.cob
ENV fprocess="./handler"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

