FROM openfaas/classic-watchdog:0.18.8 as watchdog

FROM alpine:3.11 as builder
COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /application/

RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app

RUN apk add --no-cache gcc \
    musl-dev
COPY main.c     .

RUN gcc main.c -static -o /main \
    && chmod +x /main \
    && /main

FROM scratch

COPY --from=builder /main               /
COPY --from=builder /usr/bin/fwatchdog  /

ENV fprocess="/main"
ENV suppress_lock=true

COPY --from=builder /etc/passwd /etc/passwd

USER 1000

CMD ["fwatchdog"]

