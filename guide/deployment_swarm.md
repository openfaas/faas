# Deployment guide for Docker Swarm

> Note: The best place to start is the README file in the faas or faas-netes repo.

### A foreword on security

These instructions are for a development environment. If you plan to expose OpenFaaS on the public Internet you need to enable basic authentication with a proxy such as Kong or Traefik at a minimum. TLS is also highly recomended and freely available with LetsEncrypt.org. [Kong guide](https://github.com/openfaas/faas/blob/master/guide/kong_integration.md) [Traefik guide](https://github.com/openfaas/faas/blob/master/guide/traefik_integration.md).

## Initialize Swarm Mode

You can create a single-host Docker Swarm on your laptop with a single command. You don't need any additional software to Docker 17.05 or greater. You can also run these commands on a Linux VM or cloud host.

This is how you initialize your master node:

```
# docker swarm init
```

If you have more than one IP address you may need to pass a string like `--advertise-addr eth0` to this command.

Take a note of the join token

* Join any workers you need

Log into your worker node and type in the output from `docker swarm init` on the master. If you've lost this info then type in `docker swarm join-token worker` and then enter that on the worker.

It's also important to pass the `--advertise-addr` string to any hosts which have a public IP address.

> Note: check whether you need to enable firewall rules for the [Docker Swarm ports listed here](https://docs.docker.com/engine/swarm/swarm-tutorial/).

## Deploy the stack

Clone OpenFaaS and then checkout the latest stable release:

```
$ git clone https://github.com/openfaas/faas && \
  cd faas && \
  git checkout 0.6.5 && \
  ./deploy_stack.sh
```

`./deploy_stack.sh` can be run at any time and includes a set of sample functions. You can read more about these in the [TestDrive document](https://github.com/openfaas/faas/blob/master/TestDrive.md)

## Test out the UI

Within a few seconds (or minutes if on a poor WiFi connection) the API gateway and sample functions will be pulled into your local Docker library and you will be able to access the UI at:

http://localhost:8080

If you're running on Linux you may find that `localhost` times out. In this case force an IPv4 address such as http://127.0.0.1:8080.

## Learn the CLI

You can now grab a coffee and start learning how to create your first function with the CLI:

[Morning coffee with the OpenFaaS CLI](https://blog.alexellis.io/quickstart-openfaas-cli/)
