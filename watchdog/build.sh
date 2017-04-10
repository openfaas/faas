#!/bin/sh

# Below makes use of "builder pattern" so that binary is extracted separate
# from the golang runtime/SDK

docker build --no-cache --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t functions/watchdog:build .
docker create --name buildoutput functions/watchdog:build echo
docker cp buildoutput:/go/src/github.com/alexellis/faas/watchdog/watchdog ./fwatchdog
docker rm buildoutput
