FROM arm32v7/python:2.7-slim

RUN pip install textblob && \
    python -m textblob.download_corpora

ADD https://github.com/openfaas/faas/releases/download/0.6.0/fwatchdog-armhf /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /root/

COPY handler.py .

ENV fprocess="python handler.py"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
