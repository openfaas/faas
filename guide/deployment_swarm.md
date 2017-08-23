# Deployment guide for Docker Swarm

> Note: The best place to start is the README file in the faas or faas-netes repo.

## Initialize Swarm Mode

Use either a single host or multi-node setup.

This is how you initialize your master node:

```
# docker swarm init
```

If you have more than one IP address you may need to pass a string like `--advertise-addr eth0` to this command.

Then copy any join token commands you see and run them on your worker nodes.

## Deploy the stack

```
$ git clone https://github.com/alexellis/faas && \
  cd faas && \
  git checkout 0.6.0 && \
  ./deploy_stack.sh
```

## Test the UI

Within a few seconds (or minutes if on a poor WiFi connection) the API gateway and sample functions will be pulled into your local Docker library and you will be able to access the UI at:

http://localhost:8080
