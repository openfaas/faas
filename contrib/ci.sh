#!/bin/bash

docker swarm init --advertise-addr=127.0.0.1

./deploy_stack.sh

cd ..

echo $GOPATH

mkdir -p $GOPATH/src/github.com/openfaas/
cp -r faas $GOPATH/src/github.com/openfaas/

git clone https://github.com/openfaas/certify-incubator

cp -r certify-incubator $GOPATH/src/github.com/openfaas/

cd $GOPATH/src/github.com/openfaas/faas/gateway/tests/integration && \
   go test -v

cd $GOPATH/src/github.com/openfaas/certify-incubator && \
   make test

exit 0
