# Guide on using Docker Swarm Secrets with OpenFaaS

OpenFaaS deploys functions as Docker Swarm Services, as result there are several features that we can leverage to simplify the development and subsquent deployment of functions to hardened production environments.

First an most simple is the ability to set environment variables at deploy time. For example, you might want to set the `NODE_ENV` or `DEBUG` variable.  If you are interacting with the OpenFaaS gateway via the api, seeting the `NODE_ENV` might look like this


```sh
curl -H "Content-Type: application/json" \
     -X POST \
     -d '{"service":"nodeinfo","network":"func", "image": "functions/nodehelloenv:latest", "envVars": {"NODE_ENV": "production"}}' \
     http://localhost:8080/system/functions
```

This particular function returns a simple sentence that contains the `NODE_ENV` in it.

```sh
$ curl -X POST \
       -H 'Content-Type: text/plain' \
       -H 'Content-Length: 0' \
       http://localhost:8080/function/nodehelloenv
Hello from a production machine
```

A very tempting thing to do is to now add database password or api secrets as environment variables.  However, this is not secure.  Instead, we can leverage the [Docker Swarm Secrets](https://docs.docker.com/engine/swarm/secrets/) feature to safely store and give our functions access to the needed values. Using secrets is a two step process.  Take the [ApiKeyProtected](../sample-functions/ApiKeyProtected) example function, when we deploy this function we provide a secret key that it uses to authenticate requests to it.  First we must add a secret to the swarm

```sh
docker secret create secret_api_key ~/secrets/secret_api_key.txt
```

where `~/secrets/secret_api_key.txt` is a simple text file that might look like this

```txt
R^Y$qzKzSJw51K9zP$pQ3R3N
```

Equivalently, you can pipe the value to docker via stdin like this

```sh
echo "R^Y$qzKzSJw51K9zP$pQ3R3N" | docker secret create secret_api_key -
```

Now, with the secret defined, we can deploy the function like this


```sh
curl -H "Content-Type: application/json" \
     -X POST \
     -d '{"service":"protectedapi","network":"func_functions", "image": "functions/api-key-protected:latest", "secrets": ["secret_api_key"]}' \
     http://localhost:8080/system/functions
```

Now you can test the function with these commands
```sh
$ curl -H "Content-Type: application/json" \
     -X POST \
     -H "X-Api-Key: R^Y$qzKzSJw51K9zP$pQ3R3N" \
     -d '{}' \
     http://localhost:8080/function/protectedapi

Unlocked the function!
```

```sh
$ curl -H "Content-Type: application/json" \
     -X POST \
     -H "X-Api-Key: wrong_key" \
     -d '{}' \
     http://localhost:8080/function/protectedapi

Access denied!
```


Note that unlike the `envVars` in the first example, we do not provide the secret value, just a list of names: `"secrets": ["secret_api_key"]`. The secret value has already been securely stored in the Docker swarm.  One really great result of this type of configuration is that you can simplify your function code by always referencing the same secret name, no matter the environment, the only change is how the environments are configured.