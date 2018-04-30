FROM alpine:3.7

RUN apk --update add nodejs nodejs-npm

ADD https://github.com/openfaas/faas/releases/download/0.8.0/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

COPY package.json .
COPY main.js .

RUN npm i

ENV fprocess="node main.js"
CMD ["fwatchdog"]
