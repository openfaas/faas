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

On a Mac if you're using brew then you can type in
```
$ brew install faas-cli
```

On Linux

```
$ curl -sSL https://cli.openfaas.com | sh
```

## Create a function

```
$ mkdir -p ~/functions/hello-python
$ cd ~/functions
```

*hello-python/handler.py*
```
import socket
def handle(req):
    print("Hello world from " + socket.gethostname())
```

*stack.yml*
```
provider:  
  name: faas
  gateway: http://localhost:8080

functions:  
  hello-python:
    lang: python
    handler: ./hello-python/
    image: localhost:5000/faas-hello-python
```

Let's build the function
```
$ faas-cli build -f ./stack.yml
```

Upload the function to our registry
```
$ faas-cli push -f ./stack.yml
```

Check that the image made it to the registry
```
$ curl localhost:5000/v2/faas-hello-python/tags/list
{"name":"faas-hello-python","tags":["latest"]}
```

Now we will delete the local image to be sure the deployment happens from the registry
```
$ docker rmi localhost:5000/faas-hello-python
$ docker images | grep hello | wc -l
0
```

Deploy the function from the registry
```
$ faas-cli deploy -f ./stack.yml
Deploying: hello-python.  
No existing service to remove  
Deployed.  
200 OK  
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
$ faas-cli build -f ./stack.yml && \
  faas-cli push -f ./stack.yml && \
  faas-cli deploy -f ./stack.yml
```

See that the update works
```
$ curl -X POST localhost:8080/function/hello-python
Hello EARTH from 9dacd2333c1c
```

## Scale it

Start some replicas so there are functions spanning the swarm
```
$ docker service scale hello-python=3 --detach=false
hello-python scaled to 3
overall progress: 3 out of 3 tasks
1/3: running
2/3: running
3/3: running

$ docker service ls | grep hello-python
oqvof6b9gpvl        hello-python        replicated          3/3                 localhost:5000/faas-hello-python:latest
```

Test the function across the swarm
```
$ curl -X POST localhost:8080/function/hello-python
Hello earth from 9dacd2333c1c
$ curl -X POST localhost:8080/function/hello-python
Hello earth from 281c2858c673
$ curl -X POST localhost:8080/function/hello-python
Hello earth from 9dacd2333c1c
```
