#!/bin/bash

# "openfaas/nats-queue-worker"
# ^ Already multi-arch

declare -a repos=("openfaas-incubator/openfaas-operator" "openfaas-incubator/faas-idler" "openfaas/faas" "openfaas/faas-swarm" "openfaas/faas-netes" "openfaas/faas-cli")
HERE=`pwd`
ARCH=$(uname -m)

#if [ ! -z "$CACHED" ]; then
    rm -rf staging || :
    mkdir -p staging/openfaas
    mkdir -p staging/openfaas-incubator

#fi

get_image_names() {
    if  [ "openfaas-incubator/faas-idler" = $1 ]; then
        images=("openfaas/faas-idler")
    elif  [ "openfaas/faas" = $1 ]; then
        images=("openfaas/gateway" "openfaas/basic-auth-plugin")
    elif  [ "openfaas/nats-queue-worker" = $1 ]; then
        images=("openfaas/queue-worker")
    elif  [ "openfaas-incubator/openfaas-operator" = $1 ]; then
        images=("openfaas/openfaas-operator")
    else
        images=($1)
    fi
}

if [ "$ARCH" = "armv7l" ] ; then
   ARM_VERSION="armhf"
elif [ "$ARCH" = "aarch64" ] ; then
   ARM_VERSION="arm64"
fi

echo "Target architecture: ${ARM_VERSION}"

for r in "${repos[@]}"
do
   cd $HERE

   echo -e "\nBuilding: $r\n"
   git clone https://github.com/$r ./staging/$r
   cd ./staging/$r
   pwd
   export TAG=$(git describe --abbrev=0 --tags)
   echo "Latest release: $TAG"

   get_image_names $r

   for IMAGE in "${images[@]}"
   do
      TAG_PRESENT=$(curl -s "https://hub.docker.com/v2/repositories/${IMAGE}/tags/${TAG}-${ARM_VERSION}/" | grep -Po '"message": *"[^"]*"' | grep -io 'not found')
      if [ "$TAG_PRESENT" = "not found" ]; then
      break
      fi
   done
   
   if [ "$TAG_PRESENT" = "not found" ]; then
       make ci-${ARM_VERSION}-build ci-${ARM_VERSION}-push
   else
     for IMAGE in "${images[@]}"
     do
       echo "Image is already present: ${IMAGE}:${TAG}-${ARM_VERSION}"
     done
   fi
done

echo "Docker images"

for r in "${repos[@]}"
do
   cd $HERE
   cd ./staging/$r
   export TAG=$(git describe --abbrev=0 --tags)
   echo "$r"
   get_image_names $r
   for IMAGE in "${images[@]}"
   do
   echo " ${IMAGE}:${TAG}-${ARM_VERSION}"
   done
done