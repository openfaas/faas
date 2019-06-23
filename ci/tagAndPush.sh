#!/bin/sh
set -e

IMAGE_NAME=$1
PLATFORM=""

if [ ! -z "$2" ]; then
 PLATFORM="-$2"
fi

echo "Tagging $IMAGE_NAME:$TRAVIS_TAG$PLATFORM"
docker tag $IMAGE_NAME:latest-dev$PLATFORM $IMAGE_NAME:$TRAVIS_TAG$PLATFORM;
docker tag $IMAGE_NAME:latest-dev$PLATFORM quay.io/$IMAGE_NAME:$TRAVIS_TAG$PLATFORM;

echo "Pushing $IMAGE_NAME:$TRAVIS_TAG$PLATFORM"
docker push $IMAGE_NAME:$TRAVIS_TAG$PLATFORM;
docker push quay.io/$IMAGE_NAME:$TRAVIS_TAG$PLATFORM;
