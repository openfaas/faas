#!/bin/sh

echo "Building functions/resizer:armhf..."
docker build -t functions/resizer:armhf -f Dockerfile.armhf .
