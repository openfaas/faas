# Securing OpenFaaS with Kong

[Kong](https://getkong.org) is an API gateway that provides features such as security, logging, and rate limiting. By putting this in front of OpenFaaS you can quickly get access to these things and a lot more via [the many other plugins written](https://getkong.org/plugins/) for it.

Below is a demo of how you could use Kong as an authentication layer for OpenFaaS.

## Setup OpenFaaS

```
$ docker swarm init --advertise-addr $(hostname -i)

$ git clone https://github.com/alexellis/faas && \
  cd faas && \
  ./deploy_stack.sh
```

Check that one of the sample functions works

```
$ curl localhost:8080/function/func_echoit -d 'hello world'
hello world
```

## Setup Kong
```
$ docker service create --name kong-database \
    --network func_functions --detach=false \
    -e "POSTGRES_USER=kong" \
    -e "POSTGRES_DB=kong" \
    -e "POSTGRES_PASSWORD=secretpassword" \
    postgres:9.4

$ docker service create --name=kong-migrations \
    --network func_functions --detach=false --restart-condition=none \
    -e "KONG_DATABASE=postgres" \
    -e "KONG_PG_HOST=kong-database" \
    -e "KONG_PG_PASSWORD=secretpassword" \
    kong:latest kong migrations up

$ docker service create --name kong \
    --network func_functions --detach=false \
    -e "KONG_DATABASE=postgres" \
    -e "KONG_PG_HOST=kong-database" \
    -e "KONG_PG_PASSWORD=secretpassword" \
    -e "KONG_PROXY_ACCESS_LOG=/dev/stdout" \
    -e "KONG_ADMIN_ACCESS_LOG=/dev/stdout" \
    -e "KONG_PROXY_ERROR_LOG=/dev/stderr" \
    -e "KONG_ADMIN_ERROR_LOG=/dev/stderr" \
    -p 8000:8000 \
    -p 8443:8443 \
    -p 8001:8001 \
    kong:latest
```

See that Kong us up and running
```
$ curl -i localhost:8001
HTTP/1.1 200
...
```

## Use Kong to secure OpenFaaS

Proxy OpenFaaS's functions through Kong
```
$ curl -i -X POST \
    --url http://localhost:8001/apis/ \
    --data 'name=function' \
    --data 'uris=/function' \
    --data 'upstream_url=http://gateway:8080/function'

$ curl localhost:8000/function/func_echoit -d 'hello world'
hello world
```

In order to benefit from the security Kong gives you, you should make sure only to expose Kong's public port (in this case its 8000) through your firewall. If you keep 8080 exposed, then the security Kong gives you can be circumvented.


### Require basic authentication

Enable the basic-auth plugin in Kong

```
$ curl -X POST http://localhost:8001/plugins \
    --data "name=basic-auth" \
    --data "config.hide_credentials=true"
```

Create a consumer with credentials

```
$ curl -d "username=aladdin" http://localhost:8001/consumers/

$ curl -X POST http://localhost:8001/consumers/aladdin/basic-auth \
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
$ curl -i -X POST \
    --url http://localhost:8001/apis/ \
    --data 'name=ui' \
    --data 'uris=/ui' \
    --data 'upstream_url=http://gateway:8080/ui'
```

Additionally we need to expose /system/functions since the UI makes Ajax requests to it

```
$ curl -i -X POST \
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
$ curl -i -X POST http://localhost:8001/certificates \
    -F "cert=@/tmp/selfsigned.pem" \
    -F "key=@/tmp/selfsigned.key" \
    -F "snis=example.com"

HTTP/1.1 201 Created
...
```

Use the cert to secure OpenFaaS

```
$ curl -i -X POST http://localhost:8001/apis \
    -d "name=ssl-api" \
    -d "upstream_url=http://gateway:8080" \
    -d "hosts=example.com"
HTTP/1.1 201 Created
...
```

Verify that the cert is now in use. Note the '-k' parameter is just here to work around the fact that we are using self signed certs.
```
$ curl -k https://localhost:8443/function/func_echoit \
  -d 'hello world' -H 'Authorization: Basic YWxhZGRpbjpPcGVuU2VzYW1l'
hello world
```

## Configure your firewall

Between OpenFaaS and Kong a lot of ports are exposed on your host machine. In the end it is best to only expose either 8000 or 8443 out of your network depending if you added SSL or not.

Another option is to expose both and enable [https_only](https://getkong.org/docs/0.11.x/proxy/#the-https_only-property) which is used to notify clients to upgrade to https from http.
