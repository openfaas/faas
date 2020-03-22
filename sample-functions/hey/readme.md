# Overview

[`hey` is a HTTP load generator.](https://github.com/rakyll/hey).

This is a OpenFaaS Dockerfile function, it wraps `hey`.

Use it to test your functions!

For example, generate load and watch them scale up and down!!!

## Demo

From the OpenFaas Portal:

1. Deploy the `nodeinfo` function
2. Deploy this `hey` function
3. Use `hey` to load test `nodeinfo` with a request body like this:

```$
-m POST -d '{}' http://127.0.0.1:8080/function/nodeinfo.openfaas-fn
```

## Tips

* `hey` does not randomize the request payload
* `hey` defaults Content-Type to "text/html"
* `hey` requires parameters to be in a certain order
  * For example, if you don't use the order above, the post body may be empty.
  