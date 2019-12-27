FROM openfaas/classic-watchdog:0.18.8 as watchdog

FROM alpine:3.11
RUN apk --update add nodejs nodejs-npm

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /application/

COPY package.json           .
RUN npm i
COPY handler.js             .
COPY sendColor.js           .
COPY sample_response.json   .

USER 1000

ENV fprocess="node handler.js"
CMD ["fwatchdog"]