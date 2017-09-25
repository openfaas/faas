#!/bin/bash

docker swarm init --advertise-addr=$(hostname -i)

./deploy_stack.sh

cd ..

echo $GOPATH

mkdir -p $GOPATH/go/src/github.com/openfaas/
cp -r faas $GOPATH/go/src/github.com/openfaas/

git clone https://github.com/openfaas/certify-incubator

cp -r certify-incubator $GOPATH/go/src/github.com/openfaas/

cd $GOPATH/go/src/github.com/openfaas/faas/ && \
   go test -v

cd $GOPATH/go/src/github.com/openfaas/certify-incubator && \
   make test

exit 0
