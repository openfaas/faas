#!/bin/sh

export ARCH=$(uname -m)

case "$ARCH" in

"x86_64" )              DOCKERFILE="Dockerfile"
                        eTAG="latest-dev"
                        ;;

"armv7l" )              DOCKERFILE="Dockerfile.armhf"
                        eTAG="latest-armhf-dev"
                        ;;

"armv8l" | "aarch64" )  DOCKERFILE="Dockerfile.arm64"
                        eTAG="latest-arm64-dev"
                        ;;

*) echo "Sorry, your architecture ($ARCH) cannot currently be matched to a build Dockerfile.  Please raise an issue at https://github.com/openfaas/faas"
   exit 1
   ;;
esac

echo "$1"
if [ "$1" ] ; then
  eTAG=$1
fi

echo Building functions/gateway:$eTAG

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
  -t functions/gateway:$eTAG . -f $DOCKERFILE


