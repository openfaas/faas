#!/bin/bash
set -e

if [ ! -s "$TRAVIS_TAG" ] ; then
    echo "This build will be published under the tag: ${TRAVIS_TAG}"
fi

(cd gateway && ./build.sh)
(cd watchdog && ./build.sh)
(cd auth/basic-auth && ./build.sh)