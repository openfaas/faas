#!/bin/sh
echo Building functions/gateway:build

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t functions/gateway:build . -f Dockerfile.build && \
  docker create --name gateway_extract functions/gateway:build  && \
  docker cp gateway_extract:/go/src/github.com/alexellis/faas/gateway/app ./gateway && \
  docker rm -f gateway_extract && \
echo Building functions/gateway:latest && \
docker build --no-cache -t functions/gateway:latest-dev .
