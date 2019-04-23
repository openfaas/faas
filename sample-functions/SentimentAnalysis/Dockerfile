FROM python:2.7-alpine

RUN pip install textblob && \
    python -m textblob.download_corpora

RUN apk --no-cache add curl \
    && curl -sL https://github.com/openfaas/faas/releases/download/0.13.0/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

RUN addgroup -S app \
    && adduser -S -g app app

WORKDIR /home/app

USER app
COPY requirements.txt   .
RUN pip install -r requirements.txt

RUN python -m textblob.download_corpora

COPY handler.py .
ENV fprocess="python handler.py"

HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

