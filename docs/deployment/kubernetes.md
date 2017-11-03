# Deploy OpenFaaS to Kubernetes

!!! warning "A foreword on security"
    These instructions are for a development environment. If you plan to expose OpenFaaS on the public Internet you need to enable basic authentication with a proxy such as Kong or Traefik at a minimum.

    TLS is also highly recomended and freely available from LetsEncrypt.org.

    Refer to the [Kong](https://github.com/openfaas/faas/blob/master/guide/kong_integration.md) and [Traefik](https://github.com/openfaas/faas/blob/master/guide/traefik_integration.md) Integration Guides for instructions on using them with OpenFaaS.

OpenFaaS is Kubernetes-native using *Deployments*, *Services* and *Secrets*. For more detail check out the ["faas-netes" repository](https://github.com/openfaas/faas-netes).

If you already have a working Kubernetes 1.8 cluster you can skip to the [Deploy OpenFaaS](#deploy-openfaas) section.

## Create a Kubernetes Cluster

If you do not already have a Kubernetes cluster follow this guide to deploy one so you can start evaluating OpenFaaS and building functions on your laptop or on a VM (cloud or on-prem).

* [10 minute guides for minikube / kubeadm](https://blog.alexellis.io/tag/learn-k8s/)

Additional information on [setting up Kubernetes](https://kubernetes.io/docs/setup/pick-right-solution/).

## Deploy OpenFaaS

Two alternate methods for deploying OpenFaaS on Kubernetes are provided, using Helm, and manually using `kubectl`.

If you are not familiar with Helm you should continue to the [Manual Deployment](#manual-deployment) section.

### Helm Chart

A Helm chart is provided `faas-netes` repository. Follow the link below then come back to this guide.

* [OpenFaaS Helm chart](https://github.com/openfaas/faas-netes/blob/master/HELM.md)

### Manual Deployment

Deploy either a synchronous or asynchronous OpenFaaS stack, these steps assume you are running `kubectl` on a master host.

!!! tip "Standard vs Asynchronous"
    If you're using OpenFaaS for the first time we recommend deploying the [synchronous stack](#synchronous-stack).

#### Synchronous Stack

Normal non-async OpenFaaS deployments can be carried out as follows:

1. Clone the [Faas-Netes](https://github.com/openfaas/faas-netes) repository.
    ```
    $ git clone https://github.com/openfaas/faas-netes
    ```
* Deploy the Synchnronous OpenFaaS stack.
    ```
    $ cd faas-netes
    $ kubectl apply -f ./faas.yml,monitoring.yml,rbac.yml
    ```

#### Asynchronous Stack

Alternatively OpenFaaS can be deployed with support for asynchronous invocation as follows:

!!! Note "Asynchronous Invocation"
    Asynchronous invocation works by queuing requests with [NATS](https://nats.io/) Streaming. See the [Asynchronous function guide](https://github.com/openfaas/faas/blob/master/guide/asynchronous.md) for more detail.

1. Clone the [Faas-Netes](https://github.com/openfaas/faas-netes) repository.
    ```
    $ git clone https://github.com/openfaas/faas-netes
    ```
* Deploy the OpenFaaS stack with asynchronous invocation support.
    ```
    $ cd faas-netes
    $ kubectl apply -f ./faas.async.yml,nats.yml,monitoring.yml,rbac.yml
    ```

Asynchronous invocation works by queuing requests with NATS Streaming. An alternative implementation is available with Kafka in an [open PR](https://github.com/openfaas/faas/pull/311).

!!! Tip "Further Reading"
    Asynchronous invocation works by queuing requests with [NATS](https://nats.io/) Streaming. See the [Asynchronous function guide](../developer/asynchronous.md) for more detail.


## Connect to OpenFaaS

For simplicity the default configuration uses NodePorts rather than an IngressController (which is more complicated to setup) to expose access to the OpenFaaS API Gateway and Prometheus.

| Service           | TCP port |
--------------------|----------|
| API Gateway / UI  | 31112    |
| Prometheus        | 31119    |

    ssh -L 31112:127.0.0.1:31112 -N civo@185.136.233.12 -i ~/.ssh/keys/civo-nov-2017.id_rsa

!!! tip "IngressController (Advanced Users)"
    If you're an advanced Kubernetes user, you can add an IngressController to your stack and remove the NodePort assignments.

### API Gateway

OpenFaaS should complete its deployment within a few seconds (or minutes if on a poor WiFi connection), the API gateway will be pulled into your local Docker library and you will be able to access the UI at (where `kubernetes-node-ip` is the IP address or hostname of your Kubernetes node):

* http://kubernetes-node-ip:8080

![OpenFaaS API Gateway Dashboard](https://camo.githubusercontent.com/4981b6203dfdb3c668d16326b184f7fbe0287132/68747470733a2f2f7062732e7477696d672e636f6d2f6d656469612f4337626b705a62577741416e4b73782e6a7067)

### Prometheus

The Grafana dashboard linked to OpenFaaS will be accessible at:

* http://kubernetes-node-ip:31119













### 3.0 Use OpenFaaS

After deploying OpenFaaS you can start using one of the guides or blog posts to create Serverless functions or test [community functions](https://github.com/openfaas/faas/blob/master/community.md).

![](https://camo.githubusercontent.com/72f71cb0b0f6cae1c84f5a40ad57b7a9e389d0b7/68747470733a2f2f7062732e7477696d672e636f6d2f6d656469612f44466b5575483158734141744e4a362e6a70673a6d656469756d)

You can also watch a complete walk-through of OpenFaaS on Kubernetes which demonstrates auto-scaling in action and how to use the Prometheus UI. [Video walk-through](https://www.youtube.com/watch?v=0DbrLsUvaso).

**Connect to the UI**
### Deployed


## Deploy a function

There are currently no sample functions built into this stack, but we can deploy them quickly via the UI or FaaS-CLI.

### Using the CLI

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

```yaml
provider:  
  name: faas
  gateway: http://192.168.4.95:31112
```

Now deploy the samples:
**Learn about the CLI**

```
$ faas-cli deploy -f samples.yml
```

> The `faas-cli` also supports an override of `--gateway http://...` for example:
**Build your first Python function**

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
```bash
$ git clone https://github.com/openfaas/faas-cli && \
  faas-cli deploy -f samples.yml
```

### Using the UI

The UI is exposed on NodePort 31112.

Click "New Function" and fill it out with the following:

| Field      | Value                        |
-------------|------------------------------|
| Service    | nodeinfo                     |
| Image      | functions/nodeinfo:latest    |
| fProcess   | node main.js                 |
| Network    | default                      |

## Test the function

Your function will appear after a few seconds and you can click "Invoke"

The function can also be invoked through the CLI:

```bash
$ echo -n "" | faas-cli invoke --gateway http://kubernetes-ip:31112 \
                               --name nodeinfo
$ echo -n "verbose" | faas-cli invoke --gateway http://kubernetes-ip:31112 \
                                      --name nodeinfo
```
$ echo -n "" | faas-cli invoke --gateway http://kubernetes-ip:31112 nodeinfo
$ echo -n "verbose" | faas-cli invoke --gateway http://kubernetes-ip:31112 nodeinfo
```

## Helm chart

A Helm chart is provided below with experimental support.

* [OpenFaaS Helm chart](https://github.com/openfaas/faas-netes/blob/master/HELM.md)
