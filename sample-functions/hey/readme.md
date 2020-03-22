# Overview

[`hey` is a HTTP load generator.](https://github.com/rakyll/hey).

This is a OpenFaaS Dockerfile function, it wraps `hey`.

Use it to test your functions!

For example, generate load and watch them scale up and down!!!

## Demo

1. Deploy the `nodeinfo` function
2. Deploy this function
3. Send a request like through the `hey` function to load test the `nodeinfo` function

```$
-m POST -d '{}' http://127.0.0.1:8080/function/nodeinfo.openfaas-fn
```

## Tips

* `hey` does not randomize the request payload, there's nothing saying you cannot call it a bunch with different payloads, though.
* `hey` defaults Content-Type to "text/html".
* `hey` requires parameters to be in a certain order. For example, if you don't use the order above, the post body may be empty.
