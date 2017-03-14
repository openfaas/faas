## Functions As A Service - TestDrive

FaaS is a platform for building serverless functions on Docker Swarm Mode with first class metrics. Any UNIX process can be packaged as a function in FaaS enabling you to consume a range of web events without repetitive boiler-plate coding.

#### This is a Quickstart guide for the [FaaS functions as a Service](https://github.com/alexellis/faas/) project

> A Docker stack file with a number of sample functions is provided so that you can get up and running within minutes. You can also clone the code to hack on it or package your own functions.

The guide makes use of a free testing/cloud service, but if you want to try it on your own laptop just follow the guide in the README file on Github. There is also a [blog post](http://blog.alexellis.io/functions-as-a-service/) that goes into the background of the project.

* So let's head over to http://play-with-docker.com/ and start a new session.

* Click "Add new Instance" to create a Docker host, more can be added later.

This one-shot script clones the code, initialises Docker swarm mode and then deploys the FaaS sample stack.

```
# docker swarm init --advertise-addr=$(ip addr s | grep global | grep -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\.){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])' | head -n1) && \
  git clone https://github.com/alexellis/faas && \
  cd faas && \
  ./deploy_stack.sh && \
  docker service ls
```

*The shell script makes use of a v3 docker-compose.yml file*

* Now that everything's deployed take note of the two DNS entries at the top of the screen.

![](https://pbs.twimg.com/media/C1wDi_tXUAIphu-.jpg)

#### Some of the sample functions are:

* Webhook stasher function (webhookstash) - saves webhook body into container's filesystem (Golang)
* Docker Hub Stats function (hubstats) - queries the count of images for a user on the Docker Hub (Golang)
* Node Info (nodeinfo) function - gives you the OS architecture and detailled info about the CPUS (Node.js)

#### Invoke the sample functions with curl or Postman:

Head over to the [Github repo to fork the code](https://github.com/alexellis/faas), or read on to see the input/output from the sample functions.

#### Working with the sample functions

You can access the sample functions via the command line with a HTTP POST request or by using the built-in UI portal. 

**UI portal:**

The UI portal is accessible on: http://localhost:8080/ - it show a list of functions deployed on your swarm and allows you to test them out.

View screenshot:

<a href="https://pbs.twimg.com/media/C3hDUkyWEAEgciP.jpg"><img src="https://pbs.twimg.com/media/C3hDUkyWEAEgciP.jpg" width="800"></img></a>

You can find out which services are deployed like this:

```
# docker stack ls
NAME  SERVICES
func  3

# docker stack ps func
ID            NAME               IMAGE                                  NODE  DESIRED STATE  CURRENT STATE         
rhzej73haufd  func_gateway.1     alexellis2/faas-gateway:latest         moby  Running        Running 26 minutes ago
fssz6unq3e74  func_hubstats.1    alexellis2/faas-dockerhubstats:latest  moby  Running        Running 27 minutes ago
nnlzo6u3pilg  func_prometheus.1  quay.io/prometheus/prometheus:latest   moby  Running        Running 27 minutes ago
```

* Head over to http://localhost:9090 for your Prometheus metrics

* Your function can be accessed via the gateway with curl (read on)

**Sample function: Docker Hub Stats (hubstats)**

```
# curl -X POST http://localhost:8080/function/func_hubstats -d "alexellis2"
The organisation or user alexellis2 has 99 repositories on the Docker hub.

# curl -X POST http://localhost:8080/function/func_hubstats -d "library"
The organisation or user library has 128 repositories on the Docker hub.
```

The `-d` value passes in the argument for your function. This is read via STDIN and used to query the Docker Hub to see how many images you've created/pushed.

**Sample function: webhook stasher (webhookstash)**

Another cool sample function is the Webhook Stasher which saves the body of any data posted to the service to the container's filesystem. Each file is written with the filename of the UNIX time.

```
# curl -X POST http://localhost:8080/function/func_webhookstash -d '{"event": "fork", "repo": "alexellis2/faas"}'
Webhook stashed

# docker ps|grep stash
d769ca70729d        alexellis2/faas-webhookstash@sha256:b378f1a144202baa8fb008f2c8896f5a8

# docker exec d769ca70729d find .
.
./1483999546817280727.txt
./1483999702130689245.txt
./1483999702533419188.txt
./1483999702978454621.txt
./1483999703284879767.txt
./1483999719981227578.txt
./1483999720296180414.txt
./1483999720666705381.txt
./1483999720961054638.txt
```

**Sample function: Node OS Info (nodeinfo)**

Grab OS, CPU and other info via a Node.js container using the `os` module.

```
# curl -X POST http://localhost:8080/function/func_nodeinfo -d ''

linux x64 [ { model: 'Intel(R) Xeon(R) CPU E5-2670 v2 @ 2.50GHz',
    speed: 2500,
    times: 
     { user: 3754430800,
       nice: 2450200,
       sys: 885352200,
       idle: 25599742200,
       irq: 0 } },
...
```

> Why not start the code on play-with-docker.com and then configure a Github repository to send webhook to the function?
