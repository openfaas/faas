FROM alpine:3.8
RUN apk --update add nodejs nodejs-npm

RUN apk --no-cache add curl \
    && curl -sL https://github.com/openfaas/faas/releases/download/0.9.14/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

COPY package.json   .
COPY handler.js     .
COPY parser.js   .
COPY sample.json    .

RUN npm i
ENV fprocess="node handler.js"
CMD ["fwatchdog"]
