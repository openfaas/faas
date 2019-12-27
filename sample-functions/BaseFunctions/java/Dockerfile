FROM openfaas/classic-watchdog:0.18.8 as watchdog

FROM openjdk:8u121-jdk-alpine

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /application/

COPY Handler.java .
RUN javac Handler.java

ENV fprocess="java Handler"

RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app
USER 1000

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

