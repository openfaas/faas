# Integrate Traefik with your OpenFaaS cluster

> Træfik (pronounced like traffic) is a modern HTTP reverse proxy and
> load balancer made to deploy microservices with ease.
> - traefik.io

In addition, [Traefik](https://traefik.io) offers Basic Authentication and easy SSL setup, using LetsEncrypt. It
supports several backends, such as Docker Swarm and Kubernetes.

## Setup OpenFaaS

OpenFaaS setup and configuration instructions can be found here:

* [Docker Swarm](https://github.com/alexellis/faas/blob/master/guide/deployment_swarm.md)
* [Kubernetes](https://github.com/alexellis/faas/blob/master/guide/deployment_k8s.md)

To quickly setup with Docker Swarm:
```
$ docker swarm init --advertise-addr $(hostname -i)

$ git clone https://github.com/alexellis/faas
$ cd faas
$ ./deploy_stack.sh
```

## Configure Traefik for Basic Authentication

#### Generate an MD5 hashed password

Use htpasswd to generate a hashed password
```
$  htpasswd -c ./password.txt user
```
Add a new password when prompted. The new credentials can be found in
the `password.txt` file.

#### Add Traefik configuration to docker-compose.yml

Add an entry under `services` with the Traefik configuration
```
# docker-compose.yml
version: "3.2"
services:
    traefik:
        image: traefik:v1.3
        command: -c --docker=true
            --docker.swarmmode=true
            --docker.domain=traefik
            --docker.watch=true
            --web=true
            --debug=true
            --defaultEntryPoints='http'
            --entryPoints='Name:http Address::80'
        ports:
            - 80:80
            - 8080:8080
        volumes:
            - "/var/run/docker.sock:/var/run/docker.sock"
        networks:
            - functions
        deploy:
            placement:
                constraints: [node.role == manager]
```

#### Update the Gateway service

Traefik requires some service labels to discover the gateway service.
Update the gateway configuration to remove the port property and add
the appropriate labels.
```
# docker-compose.yml
...
    gateway:
        ...
        # ports:
        #     - 8080:8080
        ...
        deploy:
            labels:
                - traefik.port=8080
                - traefik.frontend.rule=PathPrefixStrip:/openfaas
                - traefik.frontend.auth.basic=user:$$apr1$$MU....4XHRJ3. #copy/paste the contents of password.txt here
...
```
Rather than publicly exposing port 8080, the added `traefik.port` label will
make the gateway service available to Traefik on port 8080, but not
publicly. Requests will now pass through Traefik and be forwarded on. The
`PathPrefixStrip` property adds the ability to add different routes to
different services. Adding the path prefix but stripping
it as a request is passed to the appropriate service makes the `/system` and `/function` paths
available by including the `/openfaas` prefix. The `basic.auth` label should
include the username and the hashed password. Remember to escape any special
characters, so if the password contains a `$`, you can escape it by
doubling up `$$` just like the above.

#### Re-Deploy OpenFaaS

Redeploy OpenFaaS to update the service with the new changes.
```
$ ./deploy_stack.yml
```

#### Test

```
$ curl -u user:password -X POST
https://localhost/openfaas/function/func_echoit -d "hello
OpenFaaS"
hello OpenFaaS
$curl -X POST
http://localhost/openfaas/function/func_echoit -d "hello OpenFaaS"
401 Unauthorized
```
