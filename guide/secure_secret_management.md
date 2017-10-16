# Guide on using Docker Swarm Secrets with OpenFaaS

OpenFaaS deploys functions as Docker Swarm Services, as result there are several features that we can leverage to simplify the development and subsquent deployment of functions to hardened production environments.

## Using Environment Variables
First, and most simple, is the ability to set environment variables at deploy time. For example, you might want to set the `NODE_ENV` or `DEBUG` variable.  Setting the `NODE_ENV` while using the `faas-cli` might look like this

```sh
$ faas-cli deploy -f ./samples.yml
$ fass-cli invoke nodehelloenv
Hello from a production machine
```

Where your `samples.yml` stack file looks like this
```yaml
provider:
  name: faas
  gateway: http://localhost:8080

functions:
  nodehelloenv:
    lang: Dockerfile
    skip_build: true
    image: functions/nodehelloenv:latest
    environment:
        NODE_ENV: production
```


(_Optional_) If you are directly using the OpenFaaS Gateway API, then it will look like this

```sh
$ curl -H "Content-Type: application/json" \
     -X POST \
     -d '{"service":"nodeinfo","network":"func", "image": "functions/nodehelloenv:latest", "envVars": {"NODE_ENV": "production"}}' \
     http://localhost:8080/system/functions
$ curl -X POST \
       -H 'Content-Type: text/plain' \
       -H 'Content-Length: 0' \
       http://localhost:8080/function/nodehelloenv
Hello from a production machine
```


## Using Swarm Secrets
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
$ echo "R^Y$qzKzSJw51K9zP$pQ3R3N" | docker secret create secret_api_key -
$ faas-cli deploy -f ./samples.yml --secret secret_api_key
$ curl -H "Content-Type: application/json" \
     -X POST \
     -H "X-Api-Key: R^Y$qzKzSJw51K9zP$pQ3R3N" \
     -d '{}' \
     http://localhost:8080/function/protectedapi

Unlocked the function!

$ curl -H "Content-Type: application/json" \
     -X POST \
     -H "X-Api-Key: wrong_key" \
     -d '{}' \
     http://localhost:8080/function/protectedapi

Access denied!
```

Your `samples.yml` stack file looks like this
```yaml
provider:
  name: faas
  gateway: http://localhost:8080

functions:
  protectedapi:
    lang: Dockerfile
    skip_build: true
    image: functions/api-key-protected:latest
```

Note that unlike the `envVars` in the first example, we do not provide the secret value, just a list of names: `"secrets": ["secret_api_key"]`. The secret value has already been securely stored in the Docker swarm.  One really great result of this type of configuration is that you can simplify your function code by always referencing the same secret name, no matter the environment, the only change is how the environments are configured.


(_Optional_) If you are using the API directly, the above looks likes this

```sh
$ curl -H "Content-Type: application/json" \
     -X POST \
     -d '{"service":"protectedapi","network":"func_functions", "image": "functions/api-key-protected:latest", "secrets": ["secret_api_key"]}' \
     http://localhost:8080/system/functions
$ curl -H "Content-Type: application/json" \
     -X POST \
     -H "X-Api-Key: R^Y$qzKzSJw51K9zP$pQ3R3N" \
     -d '{}' \
     http://localhost:8080/function/protectedapi

Unlocked the function!
```

## Advanced Swarm Secrets
For various reasons, you might add a secret to the Swarm under a different name than you want to us in your function, e.g. if you are rotating a secret key. The Docker Swarm secret specification allows us some advanced configuration of secrets [by supplying a comma-separated value specifying the secret](https://docs.docker.com/engine/reference/commandline/service_create/#create-a-service-with-secrets).  The is best show in an example. Let's change the api key on our example function.

First add a new secret key

```sh
echo "new$qzKzSJw51K9zP$pQ3R3N" | docker secret create secret_api_key_2 -
```

Then, remove our old function and redeploy it with the new secret mounted in the same place as the old secret

```sh
$ echo "new$qzKzSJw51K9zP$pQ3R3N" | docker secret create secret_api_key_s -
$ faas-cli deploy -f ./samples.yml --secret source=secret_api_key_2,target=secret_api_key  --replace
$ curl -H "Content-Type: application/json" \
     -X POST \
     -H "X-Api-Key: new$qzKzSJw51K9zP$pQ3R3N" \
     -d '{}' \
     http://localhost:8080/function/protectedapi

Unlocked the function!
```

We reuse the sample stack file as in the previous section.


(_Optional_) Directly using the API, the above looks like this.

```sh
$ curl -H "Content-Type: applicaiton/json" \
     -X DELETE \
     -d '{"functionName": "protectedapi"}' \
     http://localhost:8080/system/functions

$ curl -H "Content-Type: application/json" \
     -X POST \
     -d '{"service":"protectedapi","network":"func_functions", "image": "functions/api-key-protected:latest", "secrets": ["source=secret_api_key_2,target=secret_api_key"]}' \
     http://localhost:8080/system/functions

$ curl -H "Content-Type: application/json" \
     -X POST \
     -H "X-Api-Key: new$qzKzSJw51K9zP$pQ3R3N" \
     -d '{}' \
     http://localhost:8080/function/protectedapi

Unlocked the function!
```