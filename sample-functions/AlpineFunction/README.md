## AlpineFunction

This is a base image for Alpine Linux which already has the watchdog added and configured with a healthcheck.

This image is published on the Docker hub as `functions/alpine:latest`.

In order to deploy it - make sure you specify an "fprocess" value at runtime i.e. `sha512sum` or `wc -l`. See the docker-compose.yml file for more details on usage.
