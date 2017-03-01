#!/bin/sh
echo Building alexellis2/faas-gateway:build-armhf

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t alexellis2/faas-gateway:build-armhf . -f Dockerfile.build.armhf

docker create --name gateway_extract alexellis2/faas-gateway:build-armhf echo
docker cp gateway_extract:/go/src/github.com/alexellis/faas/gateway/app ./gateway
docker rm -f gateway_extract

echo Building alexellis2/faas-gateway:latest-armhf-dev

docker build -t alexellis2/faas-gateway:latest-armhf-dev .
