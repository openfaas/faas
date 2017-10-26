# Guide on Asynchronous processing

By default functions are accessed synchronously via the following route:

```
$ curl --data "message" http://gateway/function/{function_name}
```

As of [PR #131](https://github.com/openfaas/faas/pull/131) asynchronous invocation is available for testing.

*Logical flow for synchronous functions:*

![](https://user-images.githubusercontent.com/6358735/29469107-cbc38c88-843e-11e7-9516-c0dd33bab63b.png)

## Why use Asynchronous processing?

* Enable longer time-outs

* Process work whenever resources are available rather than immediately

* Consume a large batch of work within a few seconds and let it process at its own pace

## How does async work?

Here is a conceptual diagram

<img width="1440" alt="screen shot 2017-10-26 at 15 55 19" src="https://user-images.githubusercontent.com/6358735/32060206-047eb75c-ba66-11e7-94d3-1343ea1811db.png">

You can also use asynchronous calls with a callback URL

<img width="1440" alt="screen shot 2017-10-26 at 15 55 06" src="https://user-images.githubusercontent.com/6358735/32060205-04545692-ba66-11e7-9e6d-b800a07b9bf5.png">

## Deploy the async stack

The reference implementation for asychronous processing uses NATS Streaming, but you are free to extend OpenFaaS and write your own [queue-worker](https://github.com/open-faas/nats-queue-worker).

Swarm:

```
$ ./deploy_extended.sh
```

K8s:

```
$ kubectl -f delete ./faas.yml
$ kubectl -f apply ./faas.async.yml,nats.yml
```

## Call a function

Functions do not need to be modified to work asynchronously, just use this alternate route:

```
$ http://gateway/async-function/{function_name}
```

If you want the function to call another function or a different endpoint when it is finished then pass the `X-Callback-Url` header. This is optional.

```
$ curl http://gateway/async-function/{function_name} --data-binary @sample.json -H "X-Callback-Url: http://gateway/function/send2slack"
```

## Extend function timeouts

Functions have three timeouts configurable by environmental variables expressed in seconds:

HTTP:

* read_timeout
* write_timeout

Hard timeout:

* exec_timeout

To make use of these just add them to your Dockerfile when needed as ENV variables.

> [Function watchdog reference](https://github.com/openfaas/faas/tree/master/watchdog)
