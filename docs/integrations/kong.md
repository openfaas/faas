# Integrate Kong with your OpenFaaS cluster

[Kong](https://getkong.org) is an API gateway that provides features such as security, logging, and rate limiting. By putting this in front of OpenFaaS you can quickly get access to these things and a lot more via [the many other plugins written](https://getkong.org/plugins/) for it.

Below is a demo of how you could use Kong as an authentication layer for OpenFaaS.

## Setup OpenFaaS

If you haven't already setup OpenFaaS then you can follow one of the deployment guides available here:

* [Docker Swarm](https://github.com/openfaas/faas/blob/master/guide/deployment_swarm.md)
* [Kubernetes](https://github.com/openfaas/faas/blob/master/guide/deployment_k8s.md)

Here is a quick reference for Docker Swarm:

```
$ docker swarm init --advertise-addr $(hostname -i)

$ git clone https://github.com/openfaas/faas && \
  cd faas && \
  ./deploy_stack.sh
```

Check that one of the sample functions works

```
$ curl localhost:8080/function/func_echoit -d 'hello world'
hello world
```

## Setup Kong

Kong stores its configuration in Postgres, so we'll create a Postgres and Kong service then run a one-off migration too.

Deploy Postgres and optionally set the `POSTGRES_PASSWORD`

```
$ docker service create --name kong-database \
    --network func_functions --detach=false \
    --constraint 'node.role == manager' \
    -e "POSTGRES_USER=kong" \
    -e "POSTGRES_DB=kong" \
    -e "POSTGRES_PASSWORD=secretpassword" \
    postgres:9.4
```

Now we will use the Kong image to populate default configuration in the Postgres database:

```
$ docker service create --name=kong-migrations \
    --network func_functions --detach=false --restart-condition=none \
    -e "KONG_DATABASE=postgres" \
    -e "KONG_PG_HOST=kong-database" \
    -e "KONG_PG_PASSWORD=secretpassword" \
    kong:latest kong migrations up
```

The last service is Kong itself:

```
$ docker service create --name kong \
    --network func_functions --detach=false \
    --constraint 'node.role == manager' \
    -e "KONG_DATABASE=postgres" \
    -e "KONG_PG_HOST=kong-database" \
    -e "KONG_PG_PASSWORD=secretpassword" \
    -e "KONG_PROXY_ACCESS_LOG=/dev/stdout" \
    -e "KONG_ADMIN_ACCESS_LOG=/dev/stdout" \
    -e "KONG_PROXY_ERROR_LOG=/dev/stderr" \
    -e "KONG_ADMIN_ERROR_LOG=/dev/stderr" \
    -p 8000:8000 \
    -p 8443:8443 \
    kong:latest
```

**Doing things the right way**

Kong has an admin port with you can expose by adding `-p 8001:8001`. In this guide we will hide the port from the off-set so that if you do not have a firewall configured yet, there is less risk of someone gaining access.

Create a `curl` command alias so we can talk to the Kong admin without exposing its ports to the network.

```
$ alias kong_admin_curl='docker exec $(docker ps -q -f name="kong\.") curl'
```
See that Kong admin is up and running
```
$ kong_admin_curl -i localhost:8001
HTTP/1.1 200
...
```

## Use Kong to secure OpenFaaS

Proxy OpenFaaS's functions through Kong
```
$ kong_admin_curl -X POST \
    --url http://localhost:8001/apis/ \
    --data 'name=function' \
    --data 'uris=/function' \
    --data 'upstream_url=http://gateway:8080/function'

$ curl localhost:8000/function/func_echoit -d 'hello world'
hello world
```

### Require basic authentication

Enable the basic-auth plugin in Kong

```
$ kong_admin_curl -X POST http://localhost:8001/plugins \
    --data "name=basic-auth" \
    --data "config.hide_credentials=true"
```

Create a consumer with credentials

```
$ kong_admin_curl -d "username=aladdin" http://localhost:8001/consumers/

$ kong_admin_curl -X POST http://localhost:8001/consumers/aladdin/basic-auth \
    --data "username=aladdin" \
    --data "password=OpenSesame"
```

Verify that authentication works

```
$ curl localhost:8000/function/func_echoit -d 'hello world'
{"message":"Unauthorized"}

$ curl localhost:8000/function/func_echoit -d 'hello world' \
    -H 'Authorization: Basic xxxxxx'
{"message":"Invalid authentication credentials"}

$ echo -n aladdin:OpenSesame | base64
YWxhZGRpbjpPcGVuU2VzYW1l

$ curl localhost:8000/function/func_echoit -d 'hello world' \
    -H 'Authorization: Basic YWxhZGRpbjpPcGVuU2VzYW1l'
hello world
```

Now lets expose the /ui directory so we can securely use the web GUI

```
$ kong_admin_curl -i -X POST \
    --url http://localhost:8001/apis/ \
    --data 'name=ui' \
    --data 'uris=/ui' \
    --data 'upstream_url=http://gateway:8080/ui'
```

Additionally we need to expose /system/functions since the UI makes Ajax requests to it

```
$ kong_admin_curl -i -X POST \
    --url http://localhost:8001/apis/ \
    --data 'name=system-functions' \
    --data 'uris=/system/functions' \
    --data 'upstream_url=http://gateway:8080/system/functions'
```

Verify that the UI is secure

```
$ curl -i localhost:8000/ui/ \
    -H 'Authorization: Basic YWxhZGRpbjpPcGVuU2VzYW1l'

HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
...
```

Now visit http://localhost:8000/ui/ in your browser where you will be asked for credentials.

### Add SSL

Basic authentication does not protect from man in the middle attacks, so lets add SSL to encrypt the communication.

Create a cert. Here in the demo, we are creating selfsigned certs, but in production you should skip this step and use your existing certificates (or get some from Lets Encrypt).
```
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /tmp/selfsigned.key -out /tmp/selfsigned.pem \
  -subj "/C=US/ST=CA/L=L/O=OrgName/OU=IT Department/CN=example.com"
```

Add cert to Kong

```
$ kong_admin_curl -X POST http://localhost:8001/certificates \
    -F "cert=$(cat /tmp/selfsigned.pem)" \
    -F "key=$(cat /tmp/selfsigned.key)" \
    -F "snis=example.com"

HTTP/1.1 201 Created
...
```

Put the cert in front OpenFaaS

```
$ kong_admin_curl -i -X POST http://localhost:8001/apis \
    -d "name=ssl-api" \
    -d "upstream_url=http://gateway:8080" \
    -d "hosts=example.com"
HTTP/1.1 201 Created
...
```

Verify that the cert is now in use. Note the '-k' parameter is just here to work around the fact that we are using self signed certs.
```
$ curl -k https://localhost:8443/function/func_echoit \
  -d 'hello world' -H 'Host: example.com '\
  -H 'Authorization: Basic YWxhZGRpbjpPcGVuU2VzYW1l'
hello world
```

## Configure your firewall

Between OpenFaaS and Kong a lot of ports are exposed on your host machine. Most importantly you should hide port 8080 since that is where OpenFaaS's functions live which you were trying to secure in the first place. In the end it is best to only expose either 8000 or 8443 out of your network depending if you added SSL or not.

Another option concerning port 8000 is to expose both 8000 and 8443 and enable [https_only](https://getkong.org/docs/latest/proxy/#the-https_only-property) which is used to notify clients to upgrade to https from http.
