FROM functions/alpine:latest

RUN apk add --no-cache nmap nmap-scripts nmap-nselibs

ENV fprocess="xargs nmap"


CMD ["fwatchdog"]
