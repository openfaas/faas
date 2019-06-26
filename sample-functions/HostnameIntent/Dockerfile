FROM functions/alpine:latest
RUN apk --update add nodejs nodejs-npm

COPY package.json   .
COPY handler.js     .
COPY sample.json    .

RUN npm i
ENV fprocess="node handler.js"
CMD ["fwatchdog"]
