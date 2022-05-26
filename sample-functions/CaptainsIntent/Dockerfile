FROM ghcr.io/openfaas/classic-watchdog:0.2.0 as watchdog

FROM alpine:3.16.0 as ship
RUN apk --update add nodejs npm

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /application/

COPY package.json   .
COPY handler.js     .
COPY parser.js   .
COPY sample.json    .

RUN npm i
ENV fprocess="node handler.js"

USER 1000

CMD ["fwatchdog"]
