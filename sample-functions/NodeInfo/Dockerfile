FROM functions/alpine:latest

USER root

RUN apk --update add nodejs nodejs-npm

COPY package.json .
COPY main.js .

RUN npm i
USER 1000

ENV fprocess="node main.js"
CMD ["fwatchdog"]
