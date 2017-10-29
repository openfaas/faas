# Deployment guide for Kubernetes

> Note: The best place to start is the README file in the faas or faas-netes repo.

This guide is for deployment to a vanilla Kubernetes 1.8 cluster running on Linux hosts. It is not a hand-book, please see the set of guides and blogs posts available at [openfaas/guide](https://github.com/openfaas/faas/tree/master/guide).

### A foreword on security

These instructions are for a development environment. If you plan to expose OpenFaaS on the public Internet you need to enable basic authentication with a proxy such as Kong or Traefik at a minimum. TLS is also highly recomended and freely available with LetsEncrypt.org. [Kong guide](https://github.com/openfaas/faas/blob/master/guide/kong_integration.md) [Traefik guide](https://github.com/openfaas/faas/blob/master/guide/traefik_integration.md).

## Kubernetes

OpenFaaS is Kubernetes-native and uses *Deployments*, *Service*s and *Secret*s. For more detail check out the ["faas-netes" repository](https://github.com/openfaas/faas-netes).

> For deploying on a cloud that supports Kubernetes *LoadBalancers* you may also want to apply the configuration in: `cloud/lb.yml`.

### 1.0 Build a cluster

You can start evaluating FaaS and building functions on your laptop or on a VM (cloud or on-prem).

* [10 minute guides for minikube / kubeadm](https://blog.alexellis.io/tag/learn-k8s/)

Additional information on [setting up Kubernetes](https://kubernetes.io/docs/setup/pick-right-solution/).

### 1.1 Pick helm or YAML files for deployment

If you'd like to use helm follow the instructions in 2.0a and then come back here, otherwise follow 2.0b to use plain `kubectl`.

### 2.0a Deploy with Helm

A Helm chart is provided `faas-netes` repository. Follow the link below then come back to this guide.

* [OpenFaaS Helm chart](https://github.com/openfaas/faas-netes/blob/master/HELM.md)

### 2.0b Deploy OpenFaaS

This step assumes you are running `kubectl` on a master host.

* Clone the code

```
$ git clone https://github.com/openfaas/faas-netes
```

Deploy a synchronous or asynchronous stack. If you're using OpenFaaS for the first time we recommend the synchronous stack. The asynchronous stack also includes NATS Streaming for queuing.

* Deploy the synchronous stack

```
$ cd faas-netes
$ kubectl apply -f ./faas.yml,monitoring.yml,rbac.yml
```

Or

* Deploy the asynchronous stack

```
$ cd faas-netes
$ kubectl apply -f ./faas.async.yml,nats.yml,monitoring.yml,rbac.yml
```

Asynchronous invocation works by queuing requests with NATS Streaming. An alternative implementation is available with Kafka in an [open PR](https://github.com/openfaas/faas/pull/311).

* See also: [Asynchronous function guide](https://github.com/openfaas/faas/blob/master/guide/asynchronous.md)

### 3.0 Use OpenFaaS

After deploying OpenFaaS you can start using one of the guides or blog posts to create Serverless functions or test [community functions](https://github.com/openfaas/faas/blob/master/community.md).

![](https://camo.githubusercontent.com/72f71cb0b0f6cae1c84f5a40ad57b7a9e389d0b7/68747470733a2f2f7062732e7477696d672e636f6d2f6d656469612f44466b5575483158734141744e4a362e6a70673a6d656469756d)

You can also watch a complete walk-through of OpenFaaS on Kubernetes which demonstrates auto-scaling in action and how to use the Prometheus UI. [Video walk-through](https://www.youtube.com/watch?v=0DbrLsUvaso).

**Connect to the UI**

For simplicity the default configuration uses NodePorts rather than an IngressController (which is more complicated to setup).

| Service           | TCP port |
--------------------|----------|
| API Gateway / UI  | 31112    |
| Prometheus        | 31119    |

> If you're an advanced Kubernetes user, you can add an IngressController to your stack and remove the NodePort assignments.

* Deploy a sample function

There are currently no sample functions built into this stack, but we can deploy them quickly via the UI or FaaS-CLI.

**Use the CLI**

* Install the CLI 

```
$ curl -sL cli.openfaas.com | sudo sh
```

Then clone some samples to deploy on your cluster.

```
$ git clone https://github.com/openfaas/faas-cli
```

Edit samples.yml and change your gateway URL from `localhost:8080` to `kubernetes-node-ip:31112`.

i.e.

```
provider:  
  name: faas
  gateway: http://192.168.4.95:31112
```

Now deploy the samples:

```
$ faas-cli deploy -f samples.yml
```

> The `faas-cli` also supports an override of `--gateway http://...` for example:

```
$ faas-cli deploy -f samples.yml --gateway http://127.0.0.1:31112
```

List the functions:

```
$ faas-cli list -f samples.yml

or

$ faas-cli list  --gateway http://127.0.0.1:31112
Function                      	Invocations    	Replicas
inception                     	0              	1    
nodejs-echo                   	0              	1    
ruby-echo                     	0              	1    
shrink-image                  	0              	1    
stronghash                    	2              	1
```

Invoke a function:

```
$ echo -n Test | faas-cli invoke stronghash --gateway http://127.0.0.1:31112
c6ee9e33cf5c6715a1d148fd73f7318884b41adcb916021e2bc0e800a5c5dd97f5142178f6ae88c8fdd98e1afb0ce4c8d2c54b5f37b30b7da1997bb33b0b8a31  -
```

* Learn about the CLI

[Morning coffee with the OpenFaaS CLI](https://blog.alexellis.io/quickstart-openfaas-cli/)

* Build your first Python function

[Your first serverless Python function with OpenFaaS](https://blog.alexellis.io/first-faas-python-function/)

**Use the UI**

The UI is exposed on NodePort 31112.

Click "New Function" and fill it out with the following:

| Field      | Value                        |
-------------|------------------------------|
| Service    | nodeinfo                     |
| Image      | functions/nodeinfo:latest    |
| fProcess   | node main.js                 |
| Network    | default                      |

* Test the function

Your function will appear after a few seconds and you can click "Invoke"

The function can also be invoked through the CLI:

```
$ echo -n "" | faas-cli invoke --gateway http://kubernetes-ip:31112 nodeinfo
$ echo -n "verbose" | faas-cli invoke --gateway http://kubernetes-ip:31112 nodeinfo
```
