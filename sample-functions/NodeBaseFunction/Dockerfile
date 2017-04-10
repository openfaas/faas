FROM node:6.9.1-alpine

ADD https://github.com/alexellis/faas/releases/download/0.5.1-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /root/

COPY package.json .

# Provides a boot-strap into your function, just add handler.js to derived image
RUN npm i
COPY faas_index.js .

ENV fprocess="node faas_index.js"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
