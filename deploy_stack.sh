#!/bin/sh

if ! [ -x "$(command -v docker)" ]; then
  echo 'Unable to find docker command, please install Docker (https://www.docker.com/) and retry' >&2
  exit 1
fi

export ARCH=$(uname -m)

case "$ARCH" in

"armv7l" | "armv6l" )            COMPOSEFILE="docker-compose.armhf.yml"
                                 ;;

"x86_64")                        COMPOSEFILE="docker-compose.yml"
                                 ;;

"aarch64" | "armv8l")           COMPOSEFILE="docker-compose.arm64.yml"
                                 ;;

*) echo "Sorry, your architecture ($ARCH) cannot currently be matched to a compose file. Please raise an issue at https://github.com/openfaas/faas"
   exit 1
   ;;
esac

if [ "$ARCH" = "x86_64" ] && [ "$1" = "extended" ] ; then
COMPOSEFILE="docker-compose.extended.yml"
fi

echo "Deploying stack for $ARCH"
docker stack deploy func --compose-file $COMPOSEFILE

