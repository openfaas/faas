FROM node:6.9.1-alpine

ADD https://github.com/openfaas/faas/releases/download/0.5.1-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /root/

COPY package.json .

RUN npm i
COPY handler.js .

ENV fprocess="node handler.js"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

