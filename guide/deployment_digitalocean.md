# Deployment guide for DigitalOcean

## Create DigitalOcean API Access Token
Follow DigitalOceans instructions to create a Personal Access Token with **Read** and **Write** permissions, give the token a descriptive name for example `openfaas-getting-started`.
- https://www.digitalocean.com/community/tutorials/how-to-use-the-digitalocean-api-v2#how-to-generate-a-personal-access-token

Set an environment variable with your token value.
```
export DOTOKEN=738cb0cd2jfhu84c33hu...
```

## Install Docker Machine
Follow Dockers instructions to install Docker Machine on your laptop/desktop.
- https://docs.docker.com/machine/install-machine/

## Create Docker Nodes
Use Docker Machine to create Docker nodes, the example below creates 3 droplets in the NYC3 zone, if you want to deploy only one Droplet change `"1 2 3"` to `"1"`.

This process will take a few minutes as Droplets are created and Docker installed.
```
for i in 1 2 3; do
    docker-machine create \
        --driver digitalocean \
        --digitalocean-image ubuntu-17-04-x64 \
        --digitalocean-tags openfaas-getting-started \
        --digitalocean-region=nyc3 \
        --digitalocean-access-token $DOTOKEN \
        node-$i;
done
```
List the newly created Docker nodes.
```
docker-machine ls
```
```
NAME     ACTIVE   DRIVER         STATE     URL                          SWARM   DOCKER        ERRORS
node-1   -        digitalocean   Running   tcp://104.131.69.233:2376            v17.07.0-ce
node-2   -        digitalocean   Running   tcp://104.131.115.146:2376           v17.07.0-ce
node-3   -        digitalocean   Running   tcp://159.203.168.121:2376           v17.07.0-ce
```

Refer to the [documentation](https://docs.docker.com/machine/drivers/digital-ocean/) for more detailed information on the DigitalOcean options for docker-machine.

# Initialize Docker Swarm
Intialize Docker Swarm on `node-1`.
```
docker-machine ssh node-1 -- docker swarm init --advertise-addr $(docker-machine ip node-1)
```
If deploying only `1` node skip to the next section.

If deploying `>1` node take a note of the command to add a worker to the Swarm. 
```
Swarm initialized: current node (je5vne1f974fea60ca75q2cac) is now a manager.

To add a worker to this swarm, run the following command:

    docker swarm join --token SWMTKN-1-239v0epdnhuol2ldguttncoaleovy29hnwyglde0kba1owc9ng-9488z5we2invwcn69f5flq7uu 104.131.69.233:2377

To add a manager to this swarm, run 'docker swarm join-token manager' and follow the instructions.
```

Add `node-2` to the Swarm, using the `docker swarm join..` command returned when initializing the master.
```
docker-machine ssh node-2 -- docker swarm join --token SWMTKN-1-239v0epdnhuol2ldguttncoaleovy29hnwyglde0kba1owc9ng-9488z5we2invwcn69f5flq7uu 104.131.69.233:2377
```
Repeat for `node-3`.
```
docker-machine ssh node-3 -- docker swarm join --token SWMTKN-1-239v0epdnhuol2ldguttncoaleovy29hnwyglde0kba1owc9ng-9488z5we2invwcn69f5flq7uu 104.131.69.233:2377
```

## Configure Docker CLI to use remote Swarm
Run this command each time you open a new shell, this tells Docker where your remote Swarm is.
```
eval $(docker-machine env node-1)
```

## Deploy the OpenFaaS Stack
```
$ git clone https://github.com/alexellis/faas && \
  cd faas && \
  git checkout 0.6.5 && \
  ./deploy_stack.sh
```

## Test the UI
Within a few seconds (or minutes if on a poor WiFi connection) the API gateway and sample functions will be deployed to the Swarm cluster running on DigitalOcean.

Access the Gateway UI via the IP address returned by `docker-machine ip node-1` (you can also access via `node-2` and `node-3`):
```
echo http://$(docker-machine ip node-1):8080
```
Prometheus metrics can be viewed at:
```
echo http://$(docker-machine ip node-1):9090
```

## Deleting OpenFaaS Droplets
You can use `docker-machine` to delete any created Droplets if are finished with your OpenFaaS deployment.
```
docker-machine rm node-1 node-2 node-3
```

## Creating a Load Balancer
Rather than address each node individually it can be preferrable to run behind a load balancer.

From the DigitalOcean console Networking page, open the Load Balancers tab and click *Create Load Balancer*.

Give the balancer a name and select the Droplets which will be balanced using the `openfaas-getting-started` tag and `NYC3` region (these were values passed to docker-machine when creating the nodes).

![create_lb](https://user-images.githubusercontent.com/83862/30240233-274c4dc0-9564-11e7-8881-54bce652392f.jpg)

Update the forwarding rules to point at the Gateway on `8080` and Prometheus dashboard on `9090`:
![forwarding_rules](https://user-images.githubusercontent.com/83862/30240106-0eb71242-9562-11e7-846e-093627026a7c.jpg)

Expand the Advanced section and update the health check to use port `8080`.
![health_checks](https://user-images.githubusercontent.com/83862/30240104-0e98e3d0-9562-11e7-89b6-c266384e35d8.jpg)

Click `Create Load Balancer` and after a few minutes your balancer will be available.

![balancer_ready](https://user-images.githubusercontent.com/83862/30240232-2747becc-9564-11e7-867a-c3ac220f2ae3.png)

You can now access the OpenFaaS Gateway, Prometheus dashboard and all functions via the load balanced IP address. For example from the balancer above:
- Gateway: http://45.55.124.29:8080
- Prometheus: http://45.55.124.29:9090
