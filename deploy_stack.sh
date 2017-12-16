#!/bin/sh

if ! [ -x "$(command -v docker)" ]; then
  echo 'Unable to find docker command, please install Docker (https://www.docker.com/) and retry' >&2
  exit 1
fi

if [ $# -eq 0 ]; then
  command="start"
else
  command="$1"
fi

case "$command" in
  start)
    echo "Deploying stack"
    docker stack deploy func --compose-file docker-compose.yml
    ;;
  stop)
    echo "Stopping stack"
    docker stack rm func
    ;;
  status)
    docker stack ps func
    ;;
  help)
    echo "Available commands:"
    echo "start  - start the stack"
    echo "stop   - stop the stack"
    echo "status - stack status"
    ;;
  *)
    echo "$0: Invalid command: $command"
    exit 1
    ;;
esac
