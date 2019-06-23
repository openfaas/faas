#!/bin/sh
set -e

export arch=$(uname -m)
export eTAG="latest-dev"

if [ "$arch" = "armv7l" ] ; then
   eTAG="latest-armhf-dev"
elif [ "$arch" = "aarch64" ] ; then
   eTAG="latest-arm64-dev"
fi

echo "$1"
if [ "$1" ] ; then
  eTAG=$1
  if [ "$arch" = "armv7l" ] ; then
    eTAG="$1-armhf"
  elif [ "$arch" = "aarch64" ] ; then
    eTAG="$1-arm64"
  fi
fi

NS=openfaas

echo Pushing $NS/basic-auth-plugin:$eTAG

docker push $NS/basic-auth-plugin:$eTAG