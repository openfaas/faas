FROM alpine:latest
RUN apk --update add nodejs
COPY ./fwatchdog /usr/bin/

COPY package.json .
COPY main.js .
RUN npm i
ENV fprocess="node main.js"
CMD ["fwatchdog"]
