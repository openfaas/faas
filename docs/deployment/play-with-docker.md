# Deployment guide for Play-with-Docker

## One-click Deployment

You can quickly start OpenFaaS on Docker Swarm online using the community-run Docker playground: play-with-docker.com (PWD) by clicking the button below:

[![Try in PWD](https://cdn.rawgit.com/play-with-docker/stacks/cff22438/assets/images/button.png)](http://play-with-docker.com?stack=https://raw.githubusercontent.com/openfaas/faas/master/docker-compose.yml&stack_name=func)

## Manual Deployment

The guide makes use of a cloud playground service called [play-with-docker.com](http://play-with-docker.com/) that provides free Docker hosts for around 5 hours. If you want to try this on your own laptop just follow along.

* Go to http://play-with-docker.com/ and start a new session. You will probably have to fill out a Captcha.

* Click "Add New Instance" to create a single Docker host (more can be added later)

This one-shot script clones the code, sets up a Docker Swarm master node then deploys OpenFaaS with the sample stack:

```
# docker swarm init --advertise-addr eth0 && \
  git clone https://github.com/openfaas/faas && \
  cd faas && \
  git checkout 0.6.7 && \
  ./deploy_stack.sh && \
  docker service ls
```

*The shell script makes use of a v3 docker-compose.yml file - read the `deploy_stack.sh` file for more details.*

* Now that everything's deployed take note of the two ports at the top of the screen:

* 8080 - the API Gateway and OpenFaaS UI
* 9090 - the Prometheus metrics endpoint

![](https://user-images.githubusercontent.com/6358735/31058899-b34f2108-a6f3-11e7-853c-6669ffacd320.jpg)
