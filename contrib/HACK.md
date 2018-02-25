# UI development

The OpenFaaS UI consists of static pages built in Angular 1.x. These call the OpenFaaS API gateway for operations such as listing / creating and deleting functions.

The [Function Store](https://github.com/openfaas/store) is stored on GitHub as a JSON file which is fecthed by the browser over HTTPS. A CORS exception is maintained for GitHub's RAW CDN for this purpose within the [gateway code](https://github.com/openfaas/faas/blob/master/gateway/server.go).

## Multi-browser testing

UI changes should be tested in:

* Safari
* Chrome
* FireFox
* IE11

### Testing on Windows

Windows VMs are available from Microsoft for free - for testing pages/projects with their browsers:
https://developer.microsoft.com/en-us/microsoft-edge/tools/vms/

[VirtualBox](https://www.virtualbox.org/wiki/Downloads) can run these VMs at no cost.

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
$ docker run --name func_gateway -e "functions_provider_url=http://faas-swarm:8080/" \
  -v `pwd`/gateway/assets:/home/app/assets \
  -v "/var/run/docker.sock:/var/run/docker.sock" \
  -p 8080:8080 --network=func_functions \
  -d functions/gateway:latest-dev
```
