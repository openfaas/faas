#!/bin/sh
echo Building functions/webhookstash:build

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy -t functions/webhookstash:build . -f Dockerfile.build && \
  docker create --name hook_extract functions/webhookstash:build
docker cp hook_extract:/go/src/app/app ./app
docker rm -f hook_extract

echo Building functions/webhookstash:latest
docker build --no-cache -t functions/webhookstash:latest .
