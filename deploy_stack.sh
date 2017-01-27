#!/bin/sh

echo "Deploying stack"
docker stack rm func ;  docker stack deploy func --compose-file docker-compose.yml 
