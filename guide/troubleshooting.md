# Troubleshooting guide

## Docker Swarm

### List all functions

```
$ docker service ls
```

### Find a function's logs

```
$ docker swarm logs --tail 100 <function>
```

### Find out if a function failed to start

```
$ docker swarm ps --no-trunc=true <function>
```

## Kubernetes

### List all functions

```
$ kubectl get deploy
```

### Find a function's logs

```
$ kubectl logs deploy/<function>
```

### Find out if a function failed to start

```
$ kubectl describe deploy/<function>
```

