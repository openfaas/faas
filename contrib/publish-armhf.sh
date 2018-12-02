#!/bin/bash

declare -a repos=("openfaas-incubator/openfaas-operator" "openfaas-incubator/faas-idler" "openfaas/faas" "openfaas/faas-swarm" "openfaas/nats-queue-worker" "openfaas/faas-netes" "openfaas/faas-cli")
HERE=`pwd`
ARCH=$(uname -m)

#if [ ! -z "$CACHED" ]; then
    rm -rf staging || :
    mkdir -p staging/openfaas
    mkdir -p staging/openfaas-incubator

#fi

get_repo_name() {
    if  [ "openfaas-incubator/faas-idler" = $1 ]; then
        echo "openfaas/faas-idler"
    elif  [ "openfaas/faas" = $1 ]; then
        echo "openfaas/gateway"
    elif  [ "openfaas/nats-queue-worker" = $1 ]; then
        echo "openfaas/queue-worker"
    elif  [ "openfaas-incubator/openfaas-operator" = $1 ]; then
        echo "openfaas/openfaas-operator"
    else
        echo $1
    fi
}

if [ "$ARCH" = "armv7l" ] ; then
   ARM_VERSION="armhf"
elif [ "$ARCH" = "aarch64" ] ; then
   ARM_VERSION="arm64"
fi

echo "Target architecture: ${ARM_VERSION}"

for i in "${repos[@]}"
do
   cd $HERE

   echo -e "\nBuilding: $i\n"
   git clone https://github.com/$i ./staging/$i
   cd ./staging/$i
   pwd
   export TAG=$(git describe --abbrev=0 --tags)
   echo "Latest release: $TAG"

   REPOSITORY=$(get_repo_name $i)
   TAG_PRESENT=$(curl -s "https://hub.docker.com/v2/repositories/${REPOSITORY}/tags/${TAG}-${ARM_VERSION}/" | grep -Po '"detail": *"[^"]*"' | grep -o 'Not found')

   if [ "$TAG_PRESENT" = "Not found" ]; then
       make ci-${ARM_VERSION}-build ci-${ARM_VERSION}-push
   else
       echo "Image is already present: ${REPOSITORY}:${TAG}-${ARM_VERSION}"
   fi
done
