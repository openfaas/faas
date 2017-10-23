FROM python:2.7-alpine

ADD ./fwatchdog /usr/bin
# ADD https://github.com/alexellis/faas/releases/download/0.5.7c-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /root/


RUN pip install flask

ENV fprocess="python handler.py"

COPY handler.py .
COPY templates templates

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

