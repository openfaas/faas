#!/bin/sh

export TAG="latest-dev"
echo Building functions/gateway:$TAG

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
  -t functions/gateway:$TAG .
