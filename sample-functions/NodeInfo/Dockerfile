FROM alpine:latest
RUN apk --update add nodejs

ADD https://github.com/alexellis/faas/releases/download/v0.3-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

COPY package.json .
COPY main.js .
RUN npm i
ENV fprocess="node main.js"
CMD ["fwatchdog"]
