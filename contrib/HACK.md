## Build a development API Gateway

Create `functions/gateway:latest-dev`

```
$ cd gateway/
$ ./build.sh
```

Now edit the gateway service in your `docker-compose.yml` file and deploy the stack.

If you want to use an overriden name then pass in the tag to the `./build.sh` script such as `./build.sh test-1`.

## Hack on the UI for the API Gateway

To hack on the UI without rebuilding the gateway mount the assets in a bind-mount like this:

Remove the Docker stack, then create the faas network as "attachable":

```
$ docker stack rm func
$ docker network create func_functions --driver=overlay --attachable=true
```

Now edit the `docker-compose.yml` file and add an attribute to the `functions` network of `attachable`.

Now you can run the gateway as its own container and bind-mount in the HTML assets.

```
$ cd faas
$ docker run -v `pwd`/gateway/assets:/root/assets -v "/var/run/docker.sock:/var/run/docker.sock" \
-p 8080:8080 --network=func_functions -ti functions/gateway:latest-dev
```

Now deploy the rest of the stack with:  `./deploy_stack.sh`.
