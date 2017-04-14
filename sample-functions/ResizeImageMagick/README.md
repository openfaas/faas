## functions/resizer

To resize an image with ImageMagick do the following:

**Deploy the resizer function**

(Make sure you have already deployed FaaS with ./deploy_stack.sh in the root of this Github repository.

* Use `curl` to deploy the function, or click *Create a new function* on the FaaS UI 
```
$ curl -s --fail localhost:8080/system/functions -d \
'{ 
   "service": "resizer",
   "image": "functions/resizer",
   "envProcess": "convert - -resize 50% fd:1",
   "network": "func_functions"
   }'
```

**Resize a picture by 50%**

Now pick an image such as the included picture of Gordon and use `curl` or a tool of your choice to send the data to the function. Pipe the result into a new file like this:

```
$ curl localhost:8080/function/resizer --data-binary @gordon.png > small_gordon.png
```

**Customize the transformation**

If you want to customise the transformation then edit the Dockerfile or the fprocess variable and create a new image.

