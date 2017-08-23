## functions/resizer

![](https://github.com/alexellis/faas/blob/master/sample-functions/ResizeImageMagick/gordon.png)

Use this FaaS function to resize an image with ImageMagick.

**Deploy the resizer function**

(Make sure you have already deployed FaaS with ./deploy_stack.sh in the root of this Github repository.

* Option 1 - click *Create a new function* on the FaaS UI

* Option 2 - use the [faas-cli](https://github.com/alexellis/faas-cli/) (experimental)

```
# curl -sSL https://get.openfaas.com | sudo sh

# faas-cli -action=deploy -image=functions/resizer -name=resizer \
  -fprocess="convert - -resize 50% fd:1"
200 OK
URL: http://localhost:8080/function/resizer
```

**Resize a picture by 50%**

Now pick an image such as the included picture of Gordon and use `curl` or a tool of your choice to send the data to the function. Pipe the result into a new file like this:

```
$ curl localhost:8080/function/resizer --data-binary @gordon.png > small_gordon.png
```

**Customize the transformation**

If you want to customise the transformation then edit the Dockerfile or the fprocess variable and create a new function.

**Remove the function**

You can remove the function with `docker service rm resizer`.

