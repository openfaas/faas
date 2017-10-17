# Deployment guide for Docker Swarm on ARM

> Note: The best place to start if you're new to OpenFaaS is the README file in the [faas](https://github.com/openfaas/faas/blob/master/README.md) or [faas-netes](https://github.com/openfaas/faas-netes/blob/master/README.md) repo.

When running OpenFaaS on ARM a key consideration is that we need to use ARM based images for our functions; this is as simple as swapping the function's Dockerfile, all other function assets remain the same.

## Initialize Swarm Mode

You can create a single-host Docker Swarm on your ARM device with a single command. You don't need any additional software to Docker 17.05 or greater.

This is how you initialize your master node:

```
# docker swarm init
```

If you have more than one IP address you may need to pass a string like `--advertise-addr eth0` to this command.

Take a note of the join token

* Join any workers you need

Log into any worker nodes and type in the output from `docker swarm init` noted earlier. If you've lost this info then type `docker swarm join-token worker` on the master and then enter the output on the worker.

It's also important to pass the `--advertise-addr` string to any hosts which have a public IP address.

> Note: check whether you need to enable firewall rules for the [Docker Swarm ports listed here](https://docs.docker.com/engine/swarm/swarm-tutorial/).

## Deploy the stack

Clone OpenFaaS and then checkout the latest stable release:

```sh
$ git clone https://github.com/openfaas/faas && \
  cd faas && \
  git checkout 0.6.5 && \
  ./deploy_stack.armhf.sh
```

`./deploy_stack.armhf.sh` can be run at any time and includes a set of sample functions. You can read more about these in the [TestDrive document](https://github.com/openfaas/faas/blob/master/TestDrive.md)

## Test out the UI

Within a few seconds (or minutes if on a poor WiFi connection) the API gateway and sample functions will be pulled into your local Docker library and you will be able to access the UI at:

http://localhost:8080

> If you find that `localhost` times out then try to force an IPv4 address such as http://127.0.0.1:8080.

## Grab the CLI

The FaaS-CLI is an OpenFaaS client through which you can build, push, deploy and invoke your functions.  One command is all you need download and install the FaaS-CLI appropriate to your architecture:

```sh
$ curl -sL https://cli.openfaas.com | sudo sh
```

To quickly test the FaaS-CLI check the version:

```sh
$ faas-cli version
```
A successful installation should yield a response similar to this:
```
   ___                   _____           ____  
  / _ \ _ __   ___ _ __ |  ___|_ _  __ _/ ___| 
 | | | | '_ \ / _ \ '_ \| |_ / _` |/ _` \___ \ 
 | |_| | |_) |  __/ | | |  _| (_| | (_| |___) |
  \___/| .__/ \___|_| |_|_|  \__,_|\__,_|____/ 
       |_|                                     

 Commit: fe0b89352e9c078c951a9f100cd4a7daae1ca15c
 Version: 0.4.18c
```  

## Using the CLI

As mentioned at the start of the guide, when running OpenFaaS on ARM a key consideration is the use of ARM based images for our functions; this is as simple as swapping the function's Dockerfile.
We'll adapt the content of an [earlier blog](https://blog.alexellis.io/quickstart-openfaas-cli/) to demonstrate.

Create a work area for your functions:
```
$ mkdir -p ~/functions && \
  cd ~/functions
```

Next use the CLI to create a new function skeleton:

```
$ faas-cli new callme --lang node
 Folder: callme created.
   ___                   _____           ____  
  / _ \ _ __   ___ _ __ |  ___|_ _  __ _/ ___| 
 | | | | '_ \ / _ \ '_ \| |_ / _` |/ _` \___ \ 
 | |_| | |_) |  __/ | | |  _| (_| | (_| |___) |
  \___/| .__/ \___|_| |_|_|  \__,_|\__,_|____/ 
       |_|                                     

 Function created in folder: callme
 Stack file written: callme.yml 
```

Now we'll have the following structure:
```sh
├── callme
│   ├── handler.js
│   └── package.json
├── callme.yml
└── template
```

Here is where the important Dockerfile 'patching' takes place.  We need to take the `node-armhf` `Dockerfile` and place it in the `node` template directory.

```
cp ./template/node-armhf/Dockerfile ./template/node/
```

With the ARM `Dockerfile` in place we are almost ready to build as we normally would.  Before we do, quickly edit `callme.yml` changing the image line to reflect your username and tag in the architecture - `alexellis/callme:armhf`

```
$ faas-cli build -f callme.yml 
 Building: callme.  
 Clearing temporary build folder: ./build/callme/  
 Preparing ./callme/ ./build/callme/function  
 Building: alexellis/callme:armhf with node template. Please wait..
 docker build -t alexellis/callme:armhf .
 Sending build context to Docker daemon  8.704kB
 Step 1/16 : FROM arm32v6/alpine:3.6
  ---> 16566b7ed19e
...

 Step 16/16 : CMD fwatchdog  
  ---> Running in 53d04c1631aa
  ---> f5e1266b0d32
 Removing intermediate container 53d04c1631aa  
 Successfully built 9dff89fae926
 Successfully tagged alexellis/callme:armhf
 Image: alexellis/callme:armhf built.
 [0] < Builder done.
```

Use the CLI to push the newly built function to Docker Hub:
```
$ faas-cli push -f callme.yml
```

## Deploy
Deploying is the act of adding the function into your OpenFaaS API gateway
```
$ faas-cli deploy -f callme.yml
 Deploying: callme.  
 No existing service to remove  
 Deployed.  
 200 OK  
 URL: http://localhost:8080/function/callme 
```
> If you find that `localhost` times out, or the response mentions `[::1]` then try to force an IPv4 address such as http://127.0.0.1:8080.

## Invoke
Test your newly deploy function on ARM:
```
$ faas-cli invoke callme
 Reading from STDIN - hit (Control + D) to stop.  
 This is my message

 {"status":"done"}
 ```
