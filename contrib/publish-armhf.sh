#!/bin/bash

declare -a repos=("openfaas-incubator/faas-idler" "openfaas/faas" "openfaas/faas-swarm" "openfaas/nats-queue-worker" "openfaas/faas-netes" "openfaas/faas-cli")

HERE=`pwd`

#if [ ! -z "$CACHED" ]; then
    rm -rf staging || :
    mkdir -p staging/openfaas
    mkdir -p staging/openfaas-incubator

#fi

for i in "${repos[@]}"
do
   cd $HERE

   echo "$i"
   git clone https://github.com/$i ./staging/$i
   cd ./staging/$i
   pwd
   export TAG=$(git describe --abbrev=0 --tags)
   echo "Latest release: $TAG"

   make ci-armhf-build ci-armhf-push

done
