#/bin/sh

dockerd &

while [ ! -e /var/run/docker.sock ] ; do sleep 0.25 && echo "Waiting for Docker socket" ; done

docker swarm init

./deploy_stack.sh

tail -f /dev/null
