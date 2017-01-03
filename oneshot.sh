#!/bin/bash

 docker network create --driver overlay --attachable functions
 git clone https://github.com/alexellis/faas && cd faas
 cd watchdog
 ./build.sh
 docker build -t catservice .
 docker service rm catservice ; docker service create --network=functions --name catservice catservice
 cd ..
 cd gateway
 docker build -t server . ;docker rm -f server; docker run -d -v /var/run/docker.sock:/var/run/docker.sock --name server -p 8080:8080 --network=functions server

