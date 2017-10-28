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

### Launch nginx

Build gwnginx from contrib directory. 

```
$ docker service rm gwnginx ; \
 docker service create --network=func_functions \
   --secret openfaas_htpasswd --publish 8081:8080 --name gwnginx gwnginx 
```


