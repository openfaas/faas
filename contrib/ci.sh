#!/bin/bash

docker swarm init --advertise-addr=127.0.0.1

./deploy_stack.sh
docker service update func_gateway --image=functions/gateway:latest-dev

# Script makes sure OpenFaaS API gateway is ready before running tests

for i in {1..30};
do
  echo "Checking if 127.0.0.1:8000 is up.. ${i}/30" 
  curl -fs 127.0.0.1:8080/

  if [ $? -eq 0 ]; then
    break
  fi
  sleep 0.5
done

cd ..

echo $GOPATH

mkdir -p $GOPATH/src/github.com/openfaas/
cp -r faas $GOPATH/src/github.com/openfaas/

git clone https://github.com/openfaas/certifier

cp -r certifier $GOPATH/src/github.com/openfaas/

cd $GOPATH/src/github.com/openfaas/faas/gateway/tests/integration && \
   go test -v

cd $GOPATH/src/github.com/openfaas/certifier && \
   make test

exit 0
