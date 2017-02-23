## Hack on the UI for the API Gateway

To hack on the UI without rebuilding the gateway mount the assets in a bind-mount like this:

```
$ docker network create func_functions --driver=overlay --attachable=true
$ docker run -v `pwd`/gateway/assets:/root/assets -v "/var/run/docker.sock:/var/run/docker.sock" \
-p 8080:8080 --network=func_functions -ti alexellis2/faas-gateway:latest-dev
```

Then edit `docker-compose.yml` to use an external network and do a `./deploy_stack.sh`.
