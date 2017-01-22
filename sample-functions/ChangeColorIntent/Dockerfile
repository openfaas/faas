FROM alpine:latest
RUN apk --update add nodejs

ADD https://github.com/alexellis/faas/releases/download/v0.1-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

COPY package.json   .
COPY handler.js     .
COPY sendColor.js   .
COPY sample.json    .

RUN npm i
ENV fprocess="node handler.js"
CMD ["fwatchdog"]
