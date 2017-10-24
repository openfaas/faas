FROM python:alpine

WORKDIR /root/
COPY requirements.txt   .
RUN pip install -r requirements.txt

ADD ./fwatchdog /usr/bin
# ADD https://github.com/alexellis/faas/releases/download/0.5.7c-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV fprocess="python handler.py"

COPY handler.py .
COPY app.py .
COPY __init__.py .
COPY templates templates
ENV cgi_body=true
ENV cgi_headers=true
ENV fast_fork=true

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

