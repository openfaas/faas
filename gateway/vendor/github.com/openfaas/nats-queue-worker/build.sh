#!/bin/sh

export eTAG="latest-dev"
echo $1
if [ $1 ] ; then
  eTAG=$1
fi

echo Building openfaas/queue-worker:$eTAG

docker build --build-arg http_proxy=$http_proxy -t openfaas/queue-worker:$eTAG .
