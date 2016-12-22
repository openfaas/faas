#!/bin/bash

docker build -t watchdog:latest . -f Dockerfile.build
docker create --name buildoutput watchdog:latest
docker cp buildoutput:/go/src/app/app ./fwatchdog
docker rm buildoutput
