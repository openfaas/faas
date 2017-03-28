#!/bin/bash

export current=$(pwd)

for dir in `ls`;
do
    test -d "$dir" || continue
    cd $dir
    echo $dir

    if [ -e ./build.sh ]
    then
        ./build.sh
    fi
    cd ..
done
