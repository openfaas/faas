#!/bin/bash
set -e

if [ ! -s "$TAG" ] ; then
    echo "This build will be published under the tag: ${TAG}"
fi

(cd gateway && ./build.sh)
(cd auth/basic-auth && ./build.sh)
