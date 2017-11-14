## Interacting with other containers and services

Here are three ways to interact with other containers or services within your Docker Swarm or Kubernetes cluster.


## Option 1 - host port

Bind a port to the host and use the host's IP / DNS-entry in the function

## Option 2 - Swarm service

If you are creating a new container such as MySQL then you can create a swarm service and simply specify the network as an additional parameter. That makes it resolvable via DNS and doesn't require a host to be exposed on the host.

```
docker service create --name redis --network=func_functions redis:latest
```
 
## Option 3 - attachable network

Go to your docker-compose YAML file and uncomment the last line which says "attachable". Now delete and re-deploy OpenFaaS (this must remove func_functions from `docker network ls`). Now when you re-create you can use `docker run --name redis --net=func_functions redis:latest`.

