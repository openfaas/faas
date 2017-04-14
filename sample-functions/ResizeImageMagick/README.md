## functions/resizer

To resize an image with ImageMagick do the following:

**Create your function on the FaaS UI or via `curl`**

```
$ curl -s --fail localhost:8080/system/functions -d '{"service": "stronghash", "image": "functions/alpine", "envProcess": "sha512sum", "network": "func_functions"}'
```


**Resize a picture by 50%**

Now pick an image such as the included picture of Gordon and use `curl` or a tool of your choice to send the data to the function. Pipe the result into a new file like this:

```
$ curl localhost:8080/function/resizer --data-binary @gordon.png > small_gordon.png
```

**Customize the transformation**

If you want to customise the transformation then edit the Dockerfile and create a new image.


