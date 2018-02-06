# Managing images

Any Docker images you build for your FaaS functions need to be accessible by any Docker Swarm worker or manager in your cluster. The only exception is during development - when you may have a single-node cluster and the image is in your local image library.

All images should be pushed to either a private repository or the Docker Hub:

* Manually built FaaS images
* Images build with the FaaS-CLI

Repositories that need authentication will require configuration when deploying your FaaS functions.

## Using private Docker registries

FaaS supports running functions from Docker images in private Docker registries.
The registry credentials can be passed on function deployment, and are then handled by Swarm for image polling.

### Deploy functions with private registries credentials

A `POST` request on `/system/function` allows you to specify private registry credentials, as a base64-encoded basic auth (user:password).
```
curl -XPOST /system/functions -d {
    "service": "functionName",
    "image": "privateregistry.domain.com/user/function",
    "envProcess": "/usr/bin/myprocess",
    "network": "func_functions",
    "registryAuth": "dXNlcjpwYXNzd29yZA=="
}
```

Base64-encoded basic auth can be resolved using your registry username and password:

````
$ echo -n "user:password" | base64
````

or (to prevent tracking in `history` command):

```
$ base64 -
(type in) user:password (hit control+D twice)
```

You can also find this string your `~/.docker/config.json` file. Docker stores credentials here after you run the `docker login` command:

```
cat ~/.docker/config.json
{
	"auths": {
		"privateregistry.domain.com": {
			"auth": "dXNlcjpwYXNzd29yZA=="
		}
	}
}
```

### Deploy your own private Docker registry

If you wish to deploy your own private registry, you can follow [Docker official documentation](https://docs.docker.com/registry/deploying/).

A quick way to get started for a private registry with TLS and authentication
is to create a VM with port 443 open to the world (for letsencrypt registration), and a registered DNS ($YOURHOST).
Then, create these two files in the current directory:

```
# docker-compose.yml
version: '2'

services:

  registry:
    restart: always
    image: registry:2
    ports:
      - 5000:5000
      - 443:5000
    environment:
      REGISTRY_AUTH: htpasswd
      REGISTRY_AUTH_HTPASSWD_PATH: /auth/htpasswd
      REGISTRY_AUTH_HTPASSWD_REALM: Registry Realm
      REGISTRY_HTTP_TLS_LETSENCRYPT_CACHEFILE: /letsencrypt/cache
      REGISTRY_HTTP_TLS_LETSENCRYPT_EMAIL: your@email.com
    volumes:
      - ./data:/var/lib/registry
      - ./auth:/auth
      - ./letsencrypt:/letsencrypt
```

```
# auth/htpasswd (generated with `docker run --entrypoint htpasswd registry:2 -Bbn testuser testpassword`)
testuser:$2y$05$Bl9siDMe7ieQHLM8e7ifaOklKrHmXymbMqfmqXs7zssj6MMGQW4le
```

Your registry is ready to be deployed by running `docker-compose up -d`.

On the client machine, you can now login and use the newly setup registry:
```
docker pull ubuntu && docker tag ubuntu $YOURHOST/ubuntu
docker login $YOURHOST # will add encoded registry credentials to ~/.docker/config.json
    Username: testuser
    Password: testpassword
docker push $YOURHOST/ubuntu
```

Images pushed to this registry can be used as functions with FaaS, provided you pass the appropriate `registryAuth` parameter at deployment time.
