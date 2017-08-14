#!/bin/sh

echo "Building functions/sentimentanalysis:armhf..."
docker build -t functions/sentimentanalysis:armhf . -f Dockerfile.armhf

