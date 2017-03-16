#!/bin/bash

(cd gateway && ./build.armhf.sh) && \
  (cd watchdog && ./build.armhf.sh)
