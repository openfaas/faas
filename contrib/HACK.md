## Build a development API Gateway

Create `functions/gateway:latest-dev`:

```
$ cd gateway/
$ ./build.sh
```

If you want to use a different tag rather than `latest-dev` then pass in the tag to `build.sh` for example:

```
$ ./build.sh test-1
```

Now edit the `docker-compose.yml` file and replace `services.gateway.image` with:

```
image: functions/gateway:latest-dev
```

Now deploy the stack with:

```
$ ./deploy_stack.sh
```

## Gateway UI development

When making changes to the gateway UI, continuously rebuilding the gateway and redeploying the stack can be tedious and time consuming. The instructions below allow development on the UI without rebuilding/redeploying by mounting the assets directory in a bind-mount.

Remove the Docker stack, then create the faas network as "attachable":

```
$ docker stack rm func
$ docker network create func_functions --driver=overlay --attachable=true
```

Now edit the `docker-compose.yml` file:

1) Replace `services.gateway.volumes` with:

```
volumes:
    - "/var/run/docker.sock:/var/run/docker.sock"
    - "./gateway/assets:/root/assets"
```

2) Replace `services.gateway.image` with:

```
image: functions/gateway:latest-dev
```

3) Replace `networks` with:

```
networks:
    functions:
        external:
            name: func_functions
```

Now deploy the stack with:

```
$ ./deploy_stack.sh
```
