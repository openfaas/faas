FROM alpine:3.7
RUN apk --update add nodejs nodejs-npm

ADD https://github.com/openfaas/faas/releases/download/0.8.0/fwatchdog /usr/bin
#COPY ./fwatchdog /usr/bin/
RUN chmod +x /usr/bin/fwatchdog

COPY package.json           .
RUN npm i
COPY handler.js             .
COPY sendColor.js           .
COPY sample_response.json   .


ENV fprocess="node handler.js"
CMD ["fwatchdog"]
