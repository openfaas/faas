#!/bin/sh
set -e

export arch=$(uname -m)
export eTAG="latest-dev"
export DOCKERFILE="Dockerfile"

if [ "$arch" = "armv7l" ] ; then
   eTAG="latest-armhf-dev"
elif [ "$arch" = "aarch64" ] ; then
   eTAG="latest-arm64-dev"
   DOCKERFILE="Dockerfile.arm64"
elif [ "$arch" = "ppc64le" ] ; then
   eTAG="latest-ppc64le-dev"
   DOCKERFILE="Dockerfile.ppc64le"
fi

echo "$1"
if [ "$1" ] ; then
  eTAG=$1
  if [ "$arch" = "armv7l" ] ; then
    eTAG="$1-armhf"
  elif [ "$arch" = "aarch64" ] ; then
    eTAG="$1-arm64"
  elif [ "$arch" = "ppc64le" ] ; then
    eTAG="$1-ppc64le"
  fi
fi

NS=openfaas

echo Building $NS/basic-auth-plugin:$eTAG

docker build -t $NS/basic-auth-plugin:$eTAG . -f $DOCKERFILE

