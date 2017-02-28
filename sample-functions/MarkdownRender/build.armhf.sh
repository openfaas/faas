#!/bin/sh
echo Building alexellis2/faas-markdownrender:build-armhf

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t alexellis2/faas-markdownrender:build-armhf \
    . -f Dockerfile.build.armhf

docker create --name render_extract alexellis2/faas-markdownrender:build-armhf
docker cp render_extract:/go/src/app/app ./app
docker rm -f render_extract

echo Building alexellis2/faas-markdownrender:latest-armhf
docker build --no-cache -t alexellis2/faas-markdownrender:latest-armhf .\
       -f Dockerfile.armhf
