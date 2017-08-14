FROM alpine:latest
RUN apk --update add nodejs nodejs-npm

ADD https://github.com/alexellis/faas/releases/download/v0.5-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

COPY package.json   .
COPY handler.js     .
COPY sample.json    .

RUN npm i
ENV fprocess="node handler.js"
CMD ["fwatchdog"]
