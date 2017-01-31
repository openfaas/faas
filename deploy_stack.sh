#!/bin/sh

echo "Deploying stack"
docker stack deploy func --compose-file docker-compose.yml 
