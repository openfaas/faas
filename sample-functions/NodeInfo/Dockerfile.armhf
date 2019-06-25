FROM openfaas/classic-watchdog:0.14.4 as watchdog

FROM alpine:3.9

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

RUN apk --update add nodejs nodejs-npm

COPY package.json .
COPY main.js .

RUN npm i

ENV fprocess="node main.js"
CMD ["fwatchdog"]
