#!/bin/sh

# First do - git clone https://github.com/alexellis/faas && cd faas

docker network create --driver overlay --attachable functions

cd watchdog
./build.sh

cp ./fwatchdog ../sample-functions/catservice/

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t alexellis2/faas-catservice .
docker service rm catservice ; docker service create --network=functions --name catservice alexellis2/faas-catservice

cd ..

cd gateway
./build.sh
docker rm -f server; docker run -d -v /var/run/docker.sock:/var/run/docker.sock --name server -p 8080:8080 --network=functions alexellis2/faas-gateway
