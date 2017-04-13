#!/bin/sh

function test_gateway {
while [ true ]
do

   curl -s localhost:8080 |grep "angular"
   if [ ! 0 -eq $? ]
   then
     echo "Gateway not ready"
     sleep 5
   else
     break
   fi
done

echo

}

function test_function_output {
while [ true ]
do
   out=$(curl -s localhost:8080/function/$1 -d "$2")
   echo $out
   if [ "$out" == "$3" ]
   then
     echo "Service $1 is ready"
     break
   else
     echo "Service $1 not ready"
     sleep 1
   fi
done

echo

}


function test_function {
while [ true ]
do
   curl localhost:8080/function/$1 -d "$2"
   if [ ! 0 -eq $? ]
   then
      echo "Service $1 not ready"
    sleep 1
   else
     echo "Service $1 is ready"
     break
   fi
done

echo

}


test_gateway

test_function func_echoit hi
test_function func_webhookstash hi
test_function func_base64 hi
test_function func_markdown "*salut*"

curl localhost:8080/system/functions -d '{"service": "stronghash", "image": "functions/alpine", "envProcess": "sha512sum", "network": "func_functions"}'

test_function_output stronghash "hi" "150a14ed5bea6cc731cf86c41566ac427a8db48ef1b9fd626664b3bfbb99071fa4c922f33dde38719b8c8354e2b7ab9d77e0e67fc12843920a712e73d558e197  -"
