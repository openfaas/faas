#!/bin/bash

export USR=$DOCKER_NS
export TAG=$TRAVIS_TAG

docker manifest create $USR/classic-watchdog:$TAG \
  openfaas/classic-watchdog:$TAG-x86_64 \
  openfaas/classic-watchdog:$TAG-armhf \
  openfaas/classic-watchdog:$TAG-arm64 \
  openfaas/classic-watchdog:$TAG-windows

docker manifest annotate $USR/classic-watchdog:$TAG --arch arm openfaas/classic-watchdog:$TAG-armhf
docker manifest annotate $USR/classic-watchdog:$TAG --arch arm64 openfaas/classic-watchdog:$TAG-arm64
docker manifest annotate $USR/classic-watchdog:$TAG --os windows openfaas/classic-watchdog:$TAG-windows

docker manifest push $USR/classic-watchdog:$TAG

