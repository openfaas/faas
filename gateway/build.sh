#!/bin/sh
set -e

export dockerfile="Dockerfile"
export arch=$(uname -m)

export eTAG="latest-dev"
export GOARM=""

if [ "$arch" = "armv7l" ] ; then
   dockerfile="Dockerfile"
   eTAG="latest-armhf-dev"
   arch="armhf"
   GOARM="7"
elif [ "$arch" = "aarch64" ] ; then
   arch="arm64"
   dockerfile="Dockerfile"
   eTAG="latest-arm64-dev"
fi

# $arch has been mutated by this point, so check for the updated values
echo "$1"
if [ "$1" ] ; then
  eTAG=$1
  if [ "$arch" = "armhf" ] ; then
    eTAG="$1-armhf"
  elif [ "$arch" = "arm64" ] ; then
    eTAG="$1-arm64"
  fi
fi

if [ "$2" ] ; then
  NS=$2
else
  NS=openfaas
fi


echo "Building $NS/gateway:$eTAG with $dockerfile for $arch"

GIT_COMMIT_MESSAGE=$(git log -1 --pretty=%B 2>&1 | head -n 1)
GIT_COMMIT_SHA=$(git rev-list -1 HEAD)
VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///' || echo dev)

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
  --build-arg GIT_COMMIT_MESSAGE="${GIT_COMMIT_MESSAGE}" --build-arg GIT_COMMIT_SHA="${GIT_COMMIT_SHA}" \
  --build-arg VERSION="${VERSION:-dev}" \
  --build-arg GOARM="${GOARM}" \
  --build-arg ARCH="${arch}" \
  -t $NS/gateway:$eTAG . -f $dockerfile
