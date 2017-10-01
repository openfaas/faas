#!/bin/bash

if [[ "$TRAVIS_TAG" != null ]]; then
    echo "This build will be published under the tag ${TRAVIS_TAG}"
fi

(cd gateway && ./build.sh)
(cd watchdog && ./build.sh)
