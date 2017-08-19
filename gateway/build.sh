#!/bin/sh

export eTAG="latest-dev"
echo $1
if [ $1 ] ; then
  echo "set this"
  eTAG=$1
fi

echo Building functions/gateway:$eTAG

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
  -t functions/gateway:$eTAG .

