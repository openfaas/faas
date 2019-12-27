FROM openfaas/classic-watchdog:0.18.8 as watchdog

FROM alpine:3.11
RUN apk --update add nodejs nodejs-npm

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
