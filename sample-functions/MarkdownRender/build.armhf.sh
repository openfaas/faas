#!/bin/sh
echo Building functions/markdownrender:build-armhf

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t functions/markdownrender:build-armhf \
    . -f Dockerfile.build.armhf

docker create --name render_extract functions/markdownrender:build-armhf
docker cp render_extract:/go/src/app/app ./app
docker rm -f render_extract

echo Building functions/markdownrender:latest-armhf
docker build --no-cache -t functions/markdownrender:latest-armhf .\
       -f Dockerfile.armhf
