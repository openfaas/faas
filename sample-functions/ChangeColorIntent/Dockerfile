FROM alpine:latest
RUN apk --update add nodejs
COPY ./fwatchdog /usr/bin/

COPY package.json   .
COPY handler.js     .
COPY sendColor.js   .
COPY sample.json    .

RUN npm i
ENV fprocess="node handler.js"
CMD ["fwatchdog"]
