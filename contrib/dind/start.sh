#!/bin/bash

docker rm -f faas
docker run --name faas --privileged -p 8080:8080 -p 9090:9090 -d alexellis2/faas-dind:0.6.5

./test.sh

echo "Quitting after 120 seconds."
sleep 120

docker rm -f faas
