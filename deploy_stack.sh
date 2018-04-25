#!/bin/sh

if ! [ -x "$(command -v docker)" ]; then
  echo 'Unable to find docker command, please install Docker (https://www.docker.com/) and retry' >&2
  exit 1
fi

echo "Deploying stack"
command='docker stack deploy func --compose-file docker-compose.yml'

if ./check_user.sh; then
  $command
else
  sudo $command
fi
