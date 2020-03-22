FROM openfaas/classic-watchdog:0.18.1 as watchdog

FROM alpine:3.11

RUN mkdir -p /home/app

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

# Add non root user
RUN addgroup -S app && adduser app -S -G app && chown app /home/app

WORKDIR /home/app

# setup hey
RUN apk --no-cache add curl && curl -o /home/app/hey https://storage.googleapis.com/hey-release/hey_linux_amd64 && chmod +x /home/app/hey

# Change from root user
USER app

# Setup some timeouts for this function...
ENV write_timeout="60"

# Run the function
ENV fprocess="xargs ./hey"
# Set to true to see request in function logs
ENV write_debug="false"

EXPOSE 8080

HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
