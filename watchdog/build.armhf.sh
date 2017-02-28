#!/bin/sh

# Below makes use of "builder pattern" so that binary is extracted separate
# from the golang runtime/SDK

echo Building alexellis2/faas-watchdog:build-armhf

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t alexellis2/faas-watchdog:build-armhf . -f Dockerfile.armhf
docker create --name buildoutput alexellis2/faas-watchdog:build-armhf echo
docker cp buildoutput:/go/src/app/app ./fwatchdog-armhf
docker rm buildoutput

