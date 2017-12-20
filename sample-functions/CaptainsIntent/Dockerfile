FROM alpine:3.6
RUN apk --update add nodejs nodejs-npm

ADD https://github.com/openfaas/faas/releases/download/0.6.9/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

COPY package.json   .
COPY handler.js     .
COPY parser.js   .
COPY sample.json    .

RUN npm i
ENV fprocess="node handler.js"
CMD ["fwatchdog"]
