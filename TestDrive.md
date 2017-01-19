#### This is a Quickstart guide for the [FaaS functions as a Service](https://github.com/alexellis/faas/) project

The guide makes use of a free testing/cloud service, but if you want to try it on your own laptop just follow the guide in the README file on Github. There is also a [blog post](http://blog.alexellis.io/functions-as-a-service/) that goes into the background of the project.

* So let's head over to http://play-with-docker.com/ and start a new session.

* Click "Add new Instance" to create a Docker host, more can be added later.

This one-shot script clones the code, initialises Docker swarm mode and then deploys the FaaS sample stack.

```
# docker swarm init --advertise-addr=$(ifconfig eth0| grep 'inet addr:'| cut -d: -f2 | awk '{ print $1}') && \
  git clone https://github.com/alexellis/faas && \
  cd faas && \
  ./deploy_stack.sh && \
  docker service ls
```

*The shell script makes use of a v3 docker-compose.yml file*

* Now that everything's deployed take note of the two DNS entries at the top of the screen.

![](https://pbs.twimg.com/media/C1wDi_tXUAIphu-.jpg)

* Head over to the README to see how to invoke the sample function for Docker Hub Stats via the `curl` commands.

#### The sample functions are:

* Webhook stasher function (webhookstash) - saves webhook body into container's filesystem (Golang)
* Docker Hub Stats function (hubstats) - queries the count of images for a user on the Docker Hub (Golang)
* Node Info (nodeinfo) function - gives you the OS architecture and detailled info about the CPUS (Node.js)

#### Invoke the sample functions with curl or Postman:

Head over to the Github repo now for the quick-start to test out the sample functions:

[Quickstart documentation with `docker-stack deploy`](https://github.com/alexellis/faas/tree/stack_1#quickstart-with-docker-stack-deploy)

Once you're up and running checkout the `gateway_functions_count` metrics on your Prometheus endpoint on *port 9090*.

#### Would you prefer a video overview?

See how to deploy FaaS onto play-with-docker.com and Docker Swarm in 1-2 minutes. See the sample functions in action and watch the graphs in Prometheus as we ramp up the amount of requests. 

* [Deep Dive into Functions as a Service (FaaS) on Docker](https://www.youtube.com/watch?v=sp1B7l5mEzc)

#### Prometheus metrics are built-in

Prometheus is built into FaaS and the sample stack, so you can check throughput for each function individually with a rate function in Prometheus like this:

![](https://pbs.twimg.com/media/C2d9IkbXAAI58fz.jpg)

