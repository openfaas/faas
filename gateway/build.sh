#!/bin/sh
echo Building alexellis2/faas-gateway:build

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t alexellis2/faas-gateway:build . -f Dockerfile.build

docker create --name gateway_extract alexellis2/faas-gateway:build 
docker cp gateway_extract:/go/src/github.com/alexellis/faas/gateway/app ./gateway
docker rm -f gateway_extract

echo Building alexellis2/faas-gateway:latest

docker build --no-cache -t alexellis2/faas-gateway:latest-dev4 .
