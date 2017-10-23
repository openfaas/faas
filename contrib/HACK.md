## Build a development API Gateway

1. Build a new development Docker image:

```
$ cd gateway/
$ ./build.sh
```

This creates a Docker image with the name `functions/gateway:latest-dev`, but if you want to use something else then pass the tag as an argument to the `./build.sh` script. I.e. `./build.sh labels-pr`.

3. Now edit the Docker image for the `gateway` service in your `docker-compose.yml` file.

4. Redeploy the stack.

Test. Repeat.

## Work on the UI the quick way

Working on the UI with the procedure above could take up to a minute to iterate between changing code and testing the changes. This section of the post shows how to bind-mount the UI assets into the API gateway as a separate container.

Remove the Docker stack, then re-define the faas network as "attachable":

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

Now deploy the rest of the stack with: `./deploy_stack.sh`.

Now you can run the gateway as its own container via `docker run` and bind-mount in the HTML assets.

```
$ docker service rm func_gateway
$ docker run --name func_gateway -v `pwd`/gateway/assets:/root/assets \
  -v "/var/run/docker.sock:/var/run/docker.sock" \
  -p 8080:8080 --network=func_functions \
  -d functions/gateway:latest-dev
```
