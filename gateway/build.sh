#!/bin/sh

export dockerfile="Dockerfile"
export arch=$(uname -m)

export eTAG="latest-dev"

if [ "$arch" = "armv7l" ] ; then
   dockerfile="Dockerfile.armhf"
   eTAG="latest-armhf-dev"
fi

echo "$1"
if [ "$1" ] ; then
  eTAG=$1
fi

echo Building functions/gateway:$eTAG

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
  -t functions/gateway:$eTAG . -f $dockerfile --no-cache


