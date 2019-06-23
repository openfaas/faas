#!/bin/sh
set -e

IMAGE_REGISTRY=$1

if [ "$IMAGE_REGISTRY" = "quay.io" ] ; then
  USERNAME=$QUAY_USERNAME
  PASSWORD=$QUAY_PASSWORD
elif [ "$IMAGE_REGISTRY" = "docker.io" ] ; then
  USERNAME=$DOCKER_USERNAME
  PASSWORD=$DOCKER_PASSWORD 
fi

echo "Attempting to log in to $IMAGE_REGISTRY"
echo $PASSWORD | docker login -u=$USERNAME --password-stdin $IMAGE_REGISTRY;
