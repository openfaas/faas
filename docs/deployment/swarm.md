# Deploy OpenFaaS to Docker Swarm

!!! warning "A Foreword On Security"
    These instructions are for a development environment. If you plan to expose OpenFaaS on the public Internet you need to enable basic authentication with a proxy such as Kong or Traefik at a minimum.

    TLS is also highly recomended and freely available from LetsEncrypt.org.

    Refer to the [Kong](https://github.com/openfaas/faas/blob/master/guide/kong_integration.md) and [Traefik](https://github.com/openfaas/faas/blob/master/guide/traefik_integration.md) Integration Guides for instructions on using them with OpenFaaS.

The deployment guide for Docker Swarm provides a simple one-line command to get you up and running in around 60 seconds.

If you already have a working Docker Swarm you can skip to the [Deploy OpenFaaS](#deploy-openfaas) section.

## Create a Docker Swarm

You can create a single-host Docker Swarm on your laptop with a single command. You don't need any additional software to Docker 17.05 or greater. You can also run these commands on a Linux VM or cloud host running Docker.

### Initialize Swarm Mode

1. Initalise the Swarm master node with:

    ```
    # docker swarm init
    ```

    !!! note "Multiple IP Addresses"
        If you have more than one IP address you may need to explicitly set the interface the Swarm will advertise on using by adding `--advertise-addr eth0` to the command above. Refer to the [Docker CLI docs](https://docs.docker.com/engine/reference/commandline/swarm_init/#--advertise-addr) for more info.

* Take a note of the join token

### Join Swarm Workers

1. Log into your worker node(s) (if any) and type in the output from `docker swarm init` on the master.

    If you've lost this info then type in `docker swarm join-token worker` and then enter that on the worker.

    It's also important to pass the `--advertise-addr` string to any hosts which have a public IP address.

    !!! note "Optional Firewall Updates"
        Check whether you need to enable firewall rules for the [Docker Swarm ports listed here](https://docs.docker.com/engine/swarm/swarm-tutorial/).

## Deploy OpenFaaS

1. Clone the OpenFaaS repo and checkout the latest stable release:

    ```
    $ git clone https://github.com/openfaas/faas && \
      cd faas
    ```

* Deploy the OpenFaaS Stack (Linux/OSX)
    ```
    $ ./deploy_stack.sh
    ```

* Deploy the OpenFaaS Stack (Windows Powershell)
    ```
    $ ./deploy_stack.ps1
    ```

`./deploy_stack.*` scripts can be run at any time and include a set of sample functions.


## Connect to OpenFaaS

### API Gateway

OpenFaaS should complete its deployment within a few seconds (or minutes if on a poor WiFi connection), the API gateway and sample functions will be pulled into your local Docker library and you will be able to access the UI at:

* http://localhost:8080

!!! tip "Localhost Times Out"
    If you're running on Linux you may find that `localhost` times out when IPv6 is enabled. In this case force an IPv4 address such as http://127.0.0.1:8080.

![OpenFaaS API Gateway Dashboard](https://camo.githubusercontent.com/4981b6203dfdb3c668d16326b184f7fbe0287132/68747470733a2f2f7062732e7477696d672e636f6d2f6d656469612f4337626b705a62577741416e4b73782e6a7067)

### Prometheus

The Grafana dashboard linked to OpenFaaS will be accessible at:

* http://localhost:9090

## Continue Getting Started

If you are following the Getting Started Guide you should proceed to [Step 2 - OpenFaaS UIs](#).
