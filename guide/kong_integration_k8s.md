# Integrate Kong with your Kubernetes OpenFaaS cluster

[Kong](https://getkong.org) is an API gateway that provides features such as security, logging, and rate limiting. By putting this in front of OpenFaaS you can quickly get access to these things and a lot more via [the many other plugins written](https://getkong.org/plugins/) for it.

Below is a demo of how you could use Kong as an authentication layer for OpenFaaS.

# Bring up Kubernetes

In this demo we will be using minikube. If you have Kubernetes and kubectl installed already you can skip this section.

The following instructions were copied from the [Kubernetes documentation](https://kubernetes.io/docs/tasks/tools/install-minikube/).

```
$ curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x ./kubectl &&	mv ./kubectl /usr/local/bin/kubectl

$ curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/

$ minikube start
```

## Setup OpenFaaS

The following instructions were copied from the [OpenFaaS Kubernetes guide](https://github.com/openfaas/faas/blob/master/guide/deployment_k8s.md).
```
$ git clone https://github.com/openfaas/faas-netes && \
    cd faas-netes && \
    kubectl apply -f ./faas.yml,monitoring.yml,rbac.yml
```

# Add a function to OpenFaaS

```
$ FAAS_GATEWAY_URL=minikube service --url gateway && \
  echo $FAAS_GATEWAY_URL
http://192.168.99.100:31112

$ kubectl run --labels="faas_function=echoit" echoit --port 8080 --image functions/catservice:latest

$ kubectl expose deployment/echoit

$ curl $FAAS_GATEWAY_URL/function/echoit -d 'hello world'
hello world
```

## Setup Kong

The following instructions were coppied from the [Kong Kubernetes guide](https://github.com/Mashape/kong-dist-kubernetes/blob/master/minikube/README.md).

Kong stores its configuration in Postgres, so we'll create a Postgres and Kong service then run a one-off migration too.

```
$ git clone https://github.com/Mashape/kong-dist-kubernetes.git

$ cd kong-dist-kubernetes/minikube/

$ kubectl create -f postgres.yaml
```

Now we will use the Kong image to populate default configuration in the Postgres database:

```
$ kubectl create -f kong_migration_postgres.yaml
```

Once job completes, you can remove the pod by running following command:

```
$ kubectl delete -f kong_migration_postgres.yaml
```
Now we can start Kong itself
```
$ kubectl create -f kong_postgres.yaml
```

Create a `curl` command alias so we can talk to the Kong admin easily.

```
$ alias kong_admin_curl='kubectl exec -it $(kubectl get pods -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | grep kong | head -1) -- curl'
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
```

Make sure we can access a function through the Kong gateway

```
$ KONG_GATEWAY_URL=minikube service --url kong-proxy && \
  echo $KONG_GATEWAY_URL
http://192.168.99.100:30726

$ curl $KONG_GATEWAY_URL/function/echoit -d 'hello world'
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
$ curl $KONG_GATEWAY_URL/function/echoit -d 'hello world'
{"message":"Unauthorized"}

$ curl $KONG_GATEWAY_URL/function/echoit -d 'hello world' \
    -H 'Authorization: Basic xxxxxx'
{"message":"Invalid authentication credentials"}

$ echo -n aladdin:OpenSesame | base64
YWxhZGRpbjpPcGVuU2VzYW1l

$ curl $KONG_GATEWAY_URL/function/echoit -d 'hello world' \
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
$ curl -i $KONG_GATEWAY_URL/ui/ \
    -H 'Authorization: Basic YWxhZGRpbjpPcGVuU2VzYW1l'

HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
...
```

Now visit http://192.168.99.100:30726/ui/ in your browser where you will be asked for credentials.

### Add SSL

Basic authentication does not protect from man in the middle attacks, so lets add SSL to encrypt the communication.

Create a cert. Here in the demo, we are creating self signed certs, but in production you should skip this step and use your existing certificates (or get some from Lets Encrypt).
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
$ KONG_GATEWAY_SSL=minikube service --url kong-proxy-ssl | sed "s/http:/https:/" && \
  echo $KONG_GATEWAY_SSL
https://192.168.99.100:30573

$ curl -k $KONG_GATEWAY_SSL/function/echoit \
  -d 'hello world' -H 'Host: example.com '\
  -H 'Authorization: Basic YWxhZGRpbjpPcGVuU2VzYW1l'
hello world
```

## Configure your firewall

At this point you will want to make sure https://192.168.99.100:30573 is forwarded through your firewall since that is where users will get access to your functions.
