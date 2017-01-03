#!/bin/sh

# Below makes use of "builder pattern" so that binary is extracted separate
# from the golang runtime/SDK

docker build -t watchdog:latest . -f Dockerfile.build
docker create --name buildoutput watchdog:latest
docker cp buildoutput:/go/src/app/app ./fwatchdog
docker rm buildoutput
