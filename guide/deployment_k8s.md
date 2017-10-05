# Deployment guide for Kubernetes

> Note: The best place to start is the README file in the faas or faas-netes repo.

## Kubernetes

OpenFaaS is Kubernetes-native and uses *Deployments* and *Services*. For more detail check out the ["faas-netes" repository](https://github.com/openfaas/faas-netes).

> For deploying on a cloud that supports Kubernetes *LoadBalancers* you may also want to apply the configuration in: `cloud/lb.yml`.

### Build a cluster

You can start evaluating FaaS and building functions on your laptop or on a VM (cloud or on-prem).

* [10 minute guides for minikube / kubeadm](https://blog.alexellis.io/tag/k8s/)

### Deploy OpenFaaS

* Clone the code

```
$ git clone https://github.com/openfaas/faas-netes
```

* Deploy the services

```
$ cd faas-netes
$ kubectl apply -f ./faas.yml,monitoring.yml,rbac.yml
```

That's it. You now have OpenFaaS deployed.

For simplicity the default configuration uses NodePorts rather than an IngressController (which is more complicated to setup).

| Service           | TCP port |
--------------------|----------|
| API Gateway / UI  | 31112    |
| Prometheus        | 31119    |

> If you're an advanced Kubernetes user, you can add an IngressController to your stack and remove the NodePort assignments.

* Deploy a sample function

There are currently no sample functions built into this stack, but we can deploy them quickly via the UI or FaaS-CLI.

**Use the CLI**

Follow the tutorials below, but change your gateway URL from `localhost:8080` to `kubernetes-node-ip:31112`

i.e.

```
provider:  
  name: faas
  gateway: http://192.168.4.95:31112
```

* Learn about the CLI

[Morning coffee with the OpenFaaS CLI](https://blog.alexellis.io/quickstart-openfaas-cli/)

* Build your first Python function

[Your first serverless Python function with OpenFaaS](https://blog.alexellis.io/first-faas-python-function/)

You can also deploy the samples from the [FaaS-cli](https://github.com/openfaas/faas-cli), but change the gateway address as above.

```
$ git clone https://github.com/openfaas/faas-cli && \
  faas-cli deploy -f samples.yml
```

**Use the UI**

Click "New Function" and fill it out with the following:

| Field      | Value                        |
-------------|------------------------------|
| Service    | nodeinfo                     |
| Image      | functions/nodeinfo:latest    |
| fProcess   | node main.js                 |
| Network    | default                      |

* Test the function

Your function will appear after a few seconds and you can click "Invoke"

You can also use the CLI like this:

```
$ echo -n "" | faas-cli invoke --gateway http://kubernetes-ip:31112 --name nodeinfo
$ echo -n "verbose" | faas-cli invoke --gateway http://kubernetes-ip:31112 --name nodeinfo
```

## Asynchronous functions

Asynchronous invocation works by queuing requests with NATS Streaminig.

Deploy the asynchronous stack like this:

```
$ cd faas-netes
$ kubectl apply -f ./faas.async.yml,nats.yml,monitoring.yml,rbac.yml
```

* See also: [Asynchronous function guide](https://github.com/openfaas/faas/blob/master/guide/asynchronous.md)

## Helm chart

A Helm chart is provided below with experimental support.

* [OpenFaaS Helm chart](https://github.com/openfaas/faas-netes/blob/master/HELM.md)
