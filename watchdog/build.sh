#!/bin/bash

docker build -t watchdog:latest . -f Dockerfile.build
docker create --name buildoutput watchdog:latest
docker cp buildoutput:/go/src/fwatchdog/fwatchdog ./
docker rm buildoutput

