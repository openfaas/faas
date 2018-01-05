## Interacting with other containers and services

Here are three ways to interact with other containers or services within your Docker Swarm or Kubernetes cluster.

### Option 1 - host port

Bind a port to the host and use the host's IP / DNS-entry in the function.

This method is agnostic to Kubernetes or Docker Swarm and works for existing appliactions which may not even use containers. It's especially useful for existing services or databases such as MySQL, redis or Mongo.

If you refer to an IP or DNS entry, it would be best to use an environmental variable in your function to configure the address.

### Option 2 - Swarm service

If you are creating a new container such as MySQL then you can create a Swarm service (or Kubernetes deployment) and simply specify the network as an additional parameter. That makes it resolvable via DNS and doesn't require a host to be exposed on the host.

```
docker service create --name redis --network=func_functions redis:latest
```

Using a service (or Kubernetes deployment) offers the advantages of orchestration and scheduling and high-availability.

### Option 3 - use an attachable network

For Docker Swarm:

Go to your docker-compose YAML file and uncomment the last line which says "attachable". Now delete and re-deploy OpenFaaS (this must remove func_functions from `docker network ls`). Now when you re-create you can use `docker run --name redis --net=func_functions redis:latest`.

This can be useful for running privileged containers such as when you need to bind-mount a Docker socket for doing builds or accessing system devices.

