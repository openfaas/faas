#!/bin/sh

# Below makes use of "builder pattern" so that binary is extracted separate
# from the golang runtime/SDK

if [ ! $http_proxy == "" ] 
then
    docker build --no-cache --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy -t functions/watchdog:build .
else
    docker build -t functions/watchdog:build .
fi

docker create --name buildoutput functions/watchdog:build echo
docker cp buildoutput:/go/src/github.com/alexellis/faas/watchdog/watchdog ./fwatchdog

docker cp buildoutput:/go/src/github.com/alexellis/faas/watchdog/watchdog-armhf ./fwatchdog-armhf
docker cp buildoutput:/go/src/github.com/alexellis/faas/watchdog/watchdog.exe ./fwatchdog.exe

docker rm buildoutput
