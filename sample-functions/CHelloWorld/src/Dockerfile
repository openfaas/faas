FROM alpine:3.6 as builder
RUN apk add --no-cache curl \
    && curl -SL https://github.com/alexellis/faas/releases/download/0.6.4/fwatchdog > /usr/bin/fwatchdog \
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

CMD ["/fwatchdog"]

