#!/bin/sh
echo Building functions/markdownrender:build

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t functions/markdownrender . -f Dockerfile.build

docker create --name render_extract functions/markdownrender
docker cp render_extract:/go/src/app/app ./app
docker rm -f render_extract

echo Building functions/markdownrender:latest
docker build --no-cache -t functions/markdownrender:latest .
