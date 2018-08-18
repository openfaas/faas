#!/bin/sh

set -e

CONTAINER=$(for f in $(docker service ps -q func_gateway); do
                docker inspect --format '{{if eq .Status.State "running"}}{{.Status.ContainerStatus.ContainerID}}{{end}}' "$f";
            done)
if [ -z "$CONTAINER" ]; then
    echo >&2 "No func_gateway container running"
    exit 1
fi

docker exec -ti "$CONTAINER" /bin/sh -c 'cat /run/secrets/basic-auth-password'
