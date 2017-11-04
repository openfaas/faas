#!/bin/sh

if ! [ -x "$(command -v docker)" ]; then
  echo 'Unable to find docker command, please install Docker (https://www.docker.com/) and retry' >&2
  exit 1
fi

echo "Create overlay network "functions" for faas"
docker network create functions -d overlay --attachable 

echo "Deploying extended stack with kafka queue"
docker stack deploy kfk --compose-file docker-compose.kafka-setup.yml
docker stack deploy func --compose-file docker-compose.kafka-queue.yml
