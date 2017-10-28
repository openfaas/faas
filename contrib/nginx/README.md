### Basic auth in 5 seconds

This guide shows you how to protect your cluster with "Basic Auth" which involves setting a username
and password. This method will prevent tampering but for production usage will also need TLS
enabling. Free TLS certificates can be generated with LetsEncrypt.

Steps:

* Generate a password file
* Push file into secret store
* Unexpose the gateway
* Create an Nginx proxy container with the new secret

* Test it out.

### Create a .htaccess:

```
$ sudo apt-get install apache2-utils
```

```
$ htpasswd -c openfaas.htpasswd admin
New password: 
Re-type new password: 
Adding password for user admin
```

Example:

```
$ cat openfaas.htpasswd 
admin:$apr1$BgwAfB5i$dfzQPXy6VliPCVqofyHsT.
```

### Create a secret in the cluster

```
$ docker secret create --label openfaas openfaas_htpasswd openfaas.htpasswd 
q70h0nsj9odbtv12vrsijcutx
```

You can now see the secret created:

```
$ docker secret ls
ID                          NAME                DRIVER              CREATED             UPDATED
q70h0nsj9odbtv12vrsijcutx   openfaas_htpasswd                       13 seconds ago      13 seconds ago
```

### Remove the exposed port on the gateway

```
$ docker service update func_gateway --publish-rm 8080
```

### Build an Nginx container (optional)

Build gwnginx from contrib directory if you need customizations.

```
$ docker build -t alexellis/gwnginx:0.1 .
```

### Launch nginx

Deploy Nginx

```
$ docker service rm gwnginx ; \
 docker service create --network=func_functions \
   --secret openfaas_htpasswd \
   --publish 8080:8080 \
   --name gwnginx alexellis/gwnginx:0.1 
```

### Connect to the UI

You can now connect to the UI on port 8080. If you use a web-browser you will be prompted for a password.

**API/CLI**

The API will require Basic Auth but can stil be used with `curl`. We have work under testing to support basic auth inside the `faas-cli` natively.


