FROM alpine:3.7 as builder
RUN apk add --no-cache curl \
    && curl -SL https://github.com/openfaas/faas/releases/download/0.9.6/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && apk --no-cache del curl

WORKDIR /root/

RUN apk add --no-cache gcc \
                       musl-dev
COPY main.c     .

RUN gcc main.c -static -o /main \
 && chmod +x /main \
 && /main

FROM scratch

COPY --from=builder /main               /
COPY --from=builder /usr/bin/fwatchdog  /
COPY --from=builder /tmp /tmp

ENV fprocess="/main"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1

CMD ["/fwatchdog"]

