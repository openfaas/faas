#!/bin/bash

docker swarm init --advertise-addr=$(hostname -i)

./deploy_stack.sh

#cd gateway/tests/integration
#go test -v
#cd ../../../

git clone https://github.com/openfaas/certify-incubator
cd certify-incubator && \
   make test

exit 0
