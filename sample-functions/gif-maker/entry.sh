#!/bin/sh
export nano=`date +%s%N`
# -s 600x400
cat - > ./$nano.mov
ffmpeg -loglevel panic -i $nano.mov -vf scale=iw*.5:ih*.5 -pix_fmt rgb24 -r 20 -f gif - | gifsicle --optimize=3 --delay=3 > /dev/stdout \
  && rm $nano.mov

