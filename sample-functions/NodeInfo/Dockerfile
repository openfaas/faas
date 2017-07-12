FROM alpine:latest
RUN apk --update add nodejs

ADD https://github.com/alexellis/faas/releases/download/0.5.6-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

COPY package.json .
COPY main.js .
RUN npm i
ENV fprocess="node main.js"
CMD ["fwatchdog"]
