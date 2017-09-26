FROM openjdk:8u121-jdk-alpine

ADD https://github.com/openfaas/faas/releases/download/0.5.1-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /root/

COPY Handler.java .
RUN javac Handler.java

ENV fprocess="java Handler"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

