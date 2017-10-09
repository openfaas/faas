# API Gateway Dashboard UI

## Build a development UI for the API Gateway

To hack on the UI without rebuilding the gateway mount the assets in a bind-mount like this:

Remove the Docker stack, then create the faas network as "attachable":

```
$ docker stack rm func
$ docker network create func_functions --driver=overlay --attachable=true
```

Now edit the `docker-compose.yml` file and replace the existing networks block with:

```
networks:
    functions:
        external:
            name: func_functions
```

Now you can run the gateway as its own container and bind-mount in the HTML assets.

```
$ docker run -v `pwd`/gateway/assets:/root/assets \
             -v "/var/run/docker.sock:/var/run/docker.sock" \
             -p 8080:8080 \
             --network=func_functions \
             -d functions/gateway:latest-dev
```

Now deploy the rest of the stack with: `./deploy_stack.sh`.
