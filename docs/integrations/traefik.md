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
                - traefik.frontend.rule=PathPrefix:/ui,/system,/function
                - traefik.frontend.auth.basic=user:$$apr1$$MU....4XHRJ3. #copy/paste the contents of password.txt here
...
```
Rather than publicly exposing port 8080, the added `traefik.port` label will
make the gateway service available to Traefik on port 8080, but not
publicly. Requests will now pass through Traefik and be forwarded on. The
`PathPrefix` property adds the ability to add different routes to
different services. Adding `/ui,/system,/function` allows for routing to all the
Gateway endpoints. The `basic.auth` label should
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
$ curl -u user:password -X POST http://localhost/function/func_echoit -d "hello OpenFaaS"
hello OpenFaaS
$curl -X POST http://localhost/function/func_echoit -d "hello OpenFaaS"
401 Unauthorized
```
Visit the browser UI at `http://localhost/ui/`. You should
be greeted with a login alert.

## Configure Traefik with SSL Support

#### Update Traefik configuration

Traefik makes it extremely easy to add SSL support using
LetsEncrypt. Add `443` to the list of ports in the `traefik`
service, add the following flags to the command property
of the `traefik` service in the `docker-compose.yml` file,
and add a new `acme` volume under the `volumes` property.
```
# docker-compose.yml
version: "3.2"
services:
    traefik:
        command: -c --docker=true
            --docker.swarmmode=true
            --docker.domain=traefik
            --docker.watch=true
            --web=true
            --debug=true
            --defaultEntryPoints='http,https'
            --acme=true
            --acme.domains='<your-domain.com, <www.your-domain-com>'
            --acme.email=your-email@email.com
            --acme.ondemand=true
            --acme.onhostrule=true
            --acme.storage=/etc/traefik/acme/acme.json
            --entryPoints='Name:https Address::443 TLS'
            --entryPoints='Name:http Address::80'
        ports:
            - 80:80
            - 8080:8080
            - 443:443
        volumes:
            - "/var/run/docker.sock:/var/run/docker.sock
            - "acme:/etc/traefik/acme"
...
```

At the bottom of the `docker-compose.yml` file, add a new
named volume.
```
volumes:
    acme:
# end of file
```

#### Re-Deploy the OpenFaaS service
```
$ ./deploy_stack.sh
```

#### Test
```
$ curl -u user:password -X POST https://your-domain.com/function/func_echoit -d "hello OpenFaaS"
hello OpenFaaS
$curl -X POST https://your-domain.com/function/func_echoit -d "hello OpenFaaS"
401 Unauthorized
```

Visit the browser UI at `https://your-domain.com/ui/`. You should
be greeted with a login alert.
