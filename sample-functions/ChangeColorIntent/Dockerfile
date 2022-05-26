FROM ghcr.io/openfaas/classic-watchdog:0.2.0 as watchdog

FROM alpine:3.16.0 as ship
RUN apk --update add nodejs npm

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