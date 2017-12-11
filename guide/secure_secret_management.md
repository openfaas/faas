# Guide on using Docker Swarm Secrets with OpenFaaS

OpenFaaS deploys functions as Docker Swarm Services, as result there are several features that we can leverage to simplify the development and subsquent deployment of functions to hardened production environments.

## Using Environment Variables
First, and least secure, is the ability to set environment variables at deploy time. For example, you might want to set the `NODE_ENV` or `DEBUG` variable.  Setting the `NODE_ENV` in the stack file `samples.yml`
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

You can then deploy and invoke the function via the `faas-cli` using
```sh
$ faas-cli deploy -f ./samples.yml
$ fass-cli invoke nodehelloenv
Hello from a production machine
```

Notice that it is using the value of `NODE_ENV` from the stack file, the default is is `dev`.


## Using Swarm Secrets
_Note_: The examples in the following section require `faas-cli` version `>=0.5.1`.

For sensitive value we can leverage the [Docker Swarm Secrets](https://docs.docker.com/engine/swarm/secrets/) feature to safely store and give our functions access to the needed values. Using secrets is a two step process.  Take the [ApiKeyProtected](../sample-functions/ApiKeyProtected) example function, when we deploy this function we provide a secret key that it uses to authenticate requests to it.  First we must add a secret to the swarm

```sh
docker secret create secret_api_key ~/secrets/secret_api_key.txt
```

where `~/secrets/secret_api_key.txt` is a simple text file that might look like this

```txt
R^YqzKzSJw51K9zPpQ3R3N
```

Equivalently, you can pipe the value to docker via stdin like this

```sh
echo "R^YqzKzSJw51K9zPpQ3R3N" | docker secret create secret_api_key -
```

Now, with the secret defined, we can deploy the function like this


```sh
$ echo "R^YqzKzSJw51K9zPpQ3R3N" | docker secret create secret_api_key -
$ faas-cli deploy -f ./samples.yml --secret secret_api_key
$ curl -H "Content-Type: application/json" \
     -X POST \
     -H "X-Api-Key: R^YqzKzSJw51K9zPpQ3R3N" \
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

## Advanced Swarm Secrets
For various reasons, you might add a secret to the Swarm under a different name than you want to use in your function, e.g. if you are rotating a secret key. The Docker Swarm secret specification allows us some advanced configuration of secrets [by supplying a comma-separated value specifying the secret](https://docs.docker.com/engine/reference/commandline/service_create/#create-a-service-with-secrets).  The is best show in an example. Let's change the api key on our example function.

First add a new secret key

```sh
echo "newqzKzSJw51K9zPpQ3R3N" | docker secret create secret_api_key_2 -
```

Then, remove our old function and redeploy it with the new secret mounted in the same place as the old secret

```sh
$ faas-cli deploy -f ./samples.yml --secret source=secret_api_key_2,target=secret_api_key  --replace
$ curl -H "Content-Type: application/json" \
     -X POST \
     -H "X-Api-Key: newqzKzSJw51K9zPpQ3R3N" \
     -d '{}' \
     http://localhost:8080/function/protectedapi

Unlocked the function!
```

We reuse the sample stack file as in the previous section.
