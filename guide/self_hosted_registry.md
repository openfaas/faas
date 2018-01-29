# Use a self-hosted registry with OpenFaaS

If you're using OpenFaaS on single host, then you don't need to push your images to a registry. They will just be used from the local Docker library.

If you are using a remote server or a multi-node cluster then you can push your function's image to a registry or the Docker Hub.

This describes how to use OpenFaaS in a swarm with your own local registry for hosting function images.

## Set up a swarm

For this example lets presume you want to create a swarm of 3 nodes. Use node1 as the manager.

This is adapted from the [Swarm deployment guide](https://github.com/openfaas/faas/blob/master/guide/deployment_swarm.md).

```
$ docker swarm init --advertise-addr $(hostname -i)
```

Now in node2 and node3 paste the output from the above command.
```
$ docker swarm join --token ...
```

## Install OpenFaaS
```
$ git clone https://github.com/openfaas/faas && \
  cd faas && \
  ./deploy_stack.sh
```

## Start a registry

Add it to the swarm and use the same network as OpenFaaS.

```
docker service create -d -p 5000:5000 \
  --network func_functions --detach=false \
  --name registry registry:2
```

Here we are using a basic local registry. You can deploy it elsewhere and use volumes depending on your persistence needs. If you would like to [use authentication with your registry, this guide may be helpful](https://github.com/openfaas/faas/blob/master/docs/managing-images.md#deploy-your-own-private-docker-registry).


## Install the CLI

This is a helper for using and deploying functions to OpenFaaS.

On Linux

```
$ curl -sSL https://cli.openfaas.com | sh
```

On a Mac if you're using brew then you can type in
```
$ brew install faas-cli
```

## Create a function

Generate function from a template

```
$ mkdir functions && cd ~/functions
$ faas-cli new hello-python --lang=python --gateway=http://localhost:8080
```

Update the print method in *hello-python/handler.py*
```
import socket
def handle(req):
    print("Hello world from " + socket.gethostname())
```

Update the image in *hello-python.yml* to read
```
    image: localhost:5000/hello-python
```

Let's build the function
```
$ faas-cli build -f hello-python.yml
```

Upload the function to our registry
```
$ faas-cli push -f hello-python.yml
```

Check that the image made it to the registry
```
$ curl localhost:5000/v2/hello-python/tags/list
{"name":"hello-python","tags":["latest"]}
```

Now we will delete the local image to be sure the deployment happens from the registry
```
$ docker rmi localhost:5000/hello-python
$ docker images | grep hello | wc -l
0
```

Deploy the function from the registry
```
$ faas-cli deploy -f hello-python.yml
Deploying: hello-python.

Deployed. 200 OK.
URL: http://localhost:8080/function/hello-python  
```

See that the function works
```
$ curl -X POST localhost:8080/function/hello-python
Hello world from 281c2858c673
```

## Update the function

hello-python/handler.py:
```
import socket
def handle(req):
    print("Hello EARTH from " + socket.gethostname())
```
Now we can rebuild, push and deploy it to the swarm.
```
$ faas-cli build -f hello-python.yml && \
  faas-cli push -f hello-python.yml && \
  faas-cli deploy -f hello-python.yml
```

See that the update works
```
$ curl -X POST localhost:8080/function/hello-python
Hello EARTH from 9dacd2333c1c
```
