#!/bin/sh
cd ./watchdog
for f in fwatchdog*; do shasum -a 256 $f > $f.sha256; done
cd ..