FROM openfaas/classic-watchdog:0.18.8 as watchdog

FROM node:6.9.1-alpine

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /application/

COPY package.json .

RUN npm install -g coffee-script && \
    npm i

COPY handler.coffee .

ENV fprocess="coffee handler.coffee"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

USER 1000

CMD ["fwatchdog"]
