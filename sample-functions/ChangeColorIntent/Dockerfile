FROM alpine:latest
RUN apk --update add nodejs nodejs-npm

ADD https://github.com/alexellis/faas/releases/download/v0.5-alpha/fwatchdog /usr/bin
#COPY ./fwatchdog /usr/bin/
RUN chmod +x /usr/bin/fwatchdog

COPY package.json           .
RUN npm i
COPY handler.js             .
COPY sendColor.js           .
COPY sample_response.json   .


ENV fprocess="node handler.js"
CMD ["fwatchdog"]
