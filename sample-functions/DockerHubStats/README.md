Building:

```
# docker service rm hubstats ; docker service create --network=functions --name=hubstats alexellis2/dockerhub-stats
```

Query the function through the gateway:

```
# curl -X POST -d "alexellis2" -v http://localhost:8080/function/hubstats
```