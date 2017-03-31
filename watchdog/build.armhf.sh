#!/bin/sh

# Below makes use of "builder pattern" so that binary is extracted separate
# from the golang runtime/SDK

echo Building functions/watchdog:build-armhf

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t functions/watchdog:build-armhf . -f Dockerfile.armhf

docker create --name buildoutput functions/watchdog:build-armhf echo

docker cp buildoutput:/go/src/github.com/alexellis/faas/watchdog/watchdog ./fwatchdog-armhf
docker rm buildoutput
