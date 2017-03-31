#!/bin/sh
echo Building functions/api-key-protected:build

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
    -t functions/api-key-protected . -f Dockerfile.build

docker create --name render_extract functions/api-key-protected
docker cp render_extract:/go/src/app/app ./app
docker rm -f render_extract

echo Building functions/api-key-protected:latest
docker build --no-cache -t functions/api-key-protected:latest .
