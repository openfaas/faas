FROM functions/alpine:latest

USER root

RUN apk add --no-cache pwgen

USER 1000

ENV fprocess="xargs pwgen -s"
