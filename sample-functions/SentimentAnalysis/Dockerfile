FROM python:2.7-alpine

RUN pip install textblob && \
    python -m textblob.download_corpora

ADD https://github.com/openfaas/faas/releases/download/0.5.1-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /root/

COPY handler.py .

ENV fprocess="python handler.py"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

