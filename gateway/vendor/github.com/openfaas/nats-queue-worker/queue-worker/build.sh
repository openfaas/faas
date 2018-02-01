#!/bin/sh

export eTAG="latest-dev"
echo $1
if [ $1 ] ; then
  eTAG=$1
fi

echo Building functions/queue-worker:$eTAG

docker build --build-arg http_proxy=$http_proxy -t functions/queue-worker:$eTAG .

