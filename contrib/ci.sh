#!/bin/bash
docker swarm init --advertise-addr=127.0.0.1
set -e

./deploy_stack.sh --no-auth

docker service update func_gateway --image=openfaas/gateway:latest-dev

# Script makes sure OpenFaaS API gateway is ready before running tests
wait_success=false
for i in {1..30};
do
  echo "Checking if 127.0.0.1:8000 is up.. ${i}/30"
  status_code=$(curl --silent --output /dev/stderr --write-out "%{http_code}" http://127.0.0.1:8080/)

  if [ "$status_code" -ge 200 -a "$status_code" -lt 400 ]; then
     echo "Deploying gateway success"
     wait_success=true
    break
  fi
  sleep 0.5
done

if [ "$wait_success" != true ] ; then
    echo "Failed to wait for gateway"
    exit 1
fi

cd ..

echo $GOPATH

if [ ! -d "$GOPATH/src/github.com/openfaas/" ]; then
    mkdir -p $GOPATH/src/github.com/openfaas/
    cp -r faas $GOPATH/src/github.com/openfaas/
fi

if [ ! -d "$GOPATH/src/github.com/openfaas/certifier" ]; then
    git clone https://github.com/openfaas/certifier
fi

echo "Deploying OpenFaaS stack.yml from $(pwd)/faas"
command -v faas-cli >/dev/null 2>&1 || curl -sSL https://cli.openfaas.com | sudo sh
faas-cli deploy -f ./faas/stack.yml

wait_success=false
for i in {1..30}
do
  echo "Checking if 127.0.0.1:8080/function/echoit is up.. ${i}/30"
  status_code=$(curl --silent --output /dev/stderr --write-out "%{http_code}" http://127.0.0.1:8080/function/echoit -d "hello")

  if [ "$status_code" -ge 200 -a "$status_code" -lt 400 ]; then
    echo "Deploying OpenFaaS stack.yml success"
    wait_success=true
    break
  else
    echo "Attempt $i lets try again"
  fi

  printf '.'
  sleep 0.5
done

if [ "$wait_success" != true ] ; then
    echo "Failed to wait for stack.yml to deploy"
    exit 1
fi

echo Running integration tests
cd $GOPATH/src/github.com/openfaas/faas/gateway/tests/integration && \
   go test -v -count=1

echo Running certifier
cd $GOPATH/src/github.com/openfaas/certifier && \
   make test

echo Integration tests all PASSED
exit 0
