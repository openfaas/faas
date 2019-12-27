FROM functions/alpine:latest

USER root

RUN apk --update add nodejs nodejs-npm

COPY package.json   .
COPY handler.js     .
COPY sample.json    .

RUN npm i

USER 1000

ENV fprocess="node handler.js"
CMD ["fwatchdog"]
