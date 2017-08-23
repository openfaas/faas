## Functions as a Service - TestDrive

FaaS is a framework for building serverless functions on Docker Swarm Mode with first class metrics. Any UNIX process can be packaged as a function in FaaS enabling you to consume a range of web events without repetitive boiler-plate coding.

Please support the project and put a **Star** on the repo.

**Overview**

A Docker stack file with a number of sample functions is provided so that you can get up and running within minutes. You can also clone the code to hack on it or package your own functions.

The guide makes use of a free testing/cloud service, but if you want to try it on your own laptop just follow the guide in the README file on Github. There is also a [blog post](http://blog.alexellis.io/functions-as-a-service/) that goes into the background of the project.

* So let's head over to http://play-with-docker.com/ and start a new session.

* Click "Add new Instance" to create a Docker host, more can be added later.

This one-shot script clones the code, initialises Docker swarm mode and then deploys the FaaS sample stack.

```
# docker swarm init --advertise-addr eth0 && \
  git clone https://github.com/alexellis/faas && \
  cd faas && \
  git checkout 0.6.1 && \
  ./deploy_stack.sh && \
  docker service ls
```

*The shell script makes use of a v3 docker-compose.yml file*

> If you are not testing on play-with-docker then remove `--advertise-addr eth0` from first line of the script.

* Now that everything's deployed take note of the two DNS entries at the top of the screen.

![](https://pbs.twimg.com/media/C1wDi_tXUAIphu-.jpg)

## Sample functions

Some of the sample functions are:

* Markdown to HTML renderer (markdownrender) - takes .MD input and produces HTML (Golang)
* Docker Hub Stats function (hubstats) - queries the count of images for a user on the Docker Hub (Golang)
* Node Info (nodeinfo) function - gives you the OS architecture and detailled info about the CPUS (Node.js)
* Webhook stasher function (webhookstash) - saves webhook body into container's filesystem - even binaries (Golang)

Newly added sample functions (not included in the sample stack)

* Hello world in - [Java/.NET core/Python/Node/Go and more](https://github.com/alexellis/faas/tree/master/sample-functions/BaseFunctions)
* [Resize images with Image Magick](https://github.com/alexellis/faas/tree/master/sample-functions/ResizeImageMagick)
* [Function protected by API key in HTTP Header](https://github.com/alexellis/faas/tree/master/sample-functions/ApiKeyProtected)

### Invoke the sample functions with curl or Postman:

Head over to the [Github and Star the project](https://github.com/alexellis/faas), or read on to see the input/output from the sample functions.

### Working with the sample functions

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
 * A saved Prometheus view is available here: [metrics overview](http://localhost:9090/graph?g0.range_input=15m&g0.expr=rate(gateway_function_invocation_total%5B20s%5D)&g0.tab=0&g1.range_input=15m&g1.expr=gateway_functions_seconds_sum+%2F+gateway_functions_seconds_counts&g1.tab=0&g2.range_input=15m&g2.expr=gateway_service_count&g2.tab=0)

* Your functions can be accessed via the gateway UI or read on for `curl`

## Building functions from templates and the CLI

The following guides show how to use the CLI and code templates to build functions.

Using a template means you only have to write a handler file in your chosen programming language such as:

* Ruby
* Node.js
* Python
* CSharp
* Or propose a template for another programming languae

Guides:

* [Your first serverless Python function with OpenFaaS](https://blog.alexellis.io/first-faas-python-function/)

* [Your first serverless .NET / C# function with OpenFaaS](https://medium.com/@rorpage/your-first-serverless-net-function-with-openfaas-27573017dedb)

## Packaging a custom Docker image

Read the developer guide:

* [Packaging a function](https://github.com/alexellis/faas/blob/master/DEV.md)

The original blog post also walks through creating a function:

* [FaaS blog post](http://blog.alexellis.io/functions-as-a-service/)

## Add new functions to FaaS at runtime

**Option 1: via the FaaS CLI**

The FaaS CLI can be used to build functions very quickly though the use of templates. See more details on the FaaS CLI [here](https://github.com/alexellis/faas-cli).

**Option 2: via FaaS UI portal**

To attach a function at runtime you can use the "Create New Function" button on the portal UI at http://localhost:8080/ 

<a href="https://pbs.twimg.com/media/C8opW3RW0AAc9Th.jpg:large"><img src="https://pbs.twimg.com/media/C8opW3RW0AAc9Th.jpg:large" width="600"></img></a>

Creating a function via the UI:

| Option                 | Usage             |
|------------------------|--------------|
| `Image`		 	| The name of the image you want to use for the function. A good starting point is functions/alpine |
| `Service Name`  	 	| Describe the name of your service. The Service Name format is: [a-zA-Z_0-9] |
| `fProcess` 		 	| The process to invoke for each function call. This must be a UNIX binary and accept input via STDIN and output via STDOUT. |
| `Network`		 	| The network `func_functions` is the default network. |

Once the create button is clicked, faas will provision a new Docker Swarm service. The newly created function will shortly be available in the list of functions on the left hand side of the UI.

**Option 3: Through docker-compose.yml stack file** 

Edit the docker-compose stack file, then run ./deploy_stack.sh - this will only update changed/added services, not existing ones.

**Option 4: Programatically through a HTTP POST to the API Gateway**

A HTTP post can also be sent via `curl` etc to the endpoint used by the UI (HTTP post to `/system/functions`)

```
// CreateFunctionRequest create a function in the swarm.
type CreateFunctionRequest struct {
	Service    string `json:"service"`
	Image      string `json:"image"`
	Network    string `json:"network"`
	EnvProcess string `json:"envProcess"`
}
```

Example:

For a quote-of-the-day type of application:

```
curl localhost:8080/system/functions -d '
{"service": "oblique", "image": "vielmetti/faas-oblique", "envProcess": "/usr/bin/oblique", "network": "func_functions"}'
```

For a hashing algorithm:

```
curl localhost:8080/system/functions -d '
{"service": "stronghash", "image": "functions/alpine", "envProcess": "sha512sum", "network": "func_functions"}'
```

### Delete a function at runtime

You can delete a function through the FaaS-CLI or with the Docker CLI

```
$ docker service rm func_echoit
```

### Exploring the functions with `curl`

**Sample function: Docker Hub Stats (hubstats)**

```
# curl -X POST http://localhost:8080/function/func_hubstats -d "alexellis2"
The organisation or user alexellis2 has 99 repositories on the Docker hub.

# curl -X POST http://localhost:8080/function/func_hubstats -d "library"
The organisation or user library has 128 repositories on the Docker hub.
```

The `-d` value passes in the argument for your function. This is read via STDIN and used to query the Docker Hub to see how many images you've created/pushed.


**Sample function: Node OS Info (nodeinfo)**

Grab OS, CPU and other info via a Node.js container using the `os` module.

If you invoke this method in a while loop or with a load-generator tool then it will auto-scale to 5, 10, 15 and finally 20 replicas due to the load. You will then be able to see the various Docker containers responding with a different Hostname for each request as the work is distributed evenly.

Here is a loop that can be used to invoke the function in a loop to trigger auto-scaling.
```
while [ true ] ; do curl -X POST http://localhost:8080/function/func_nodeinfo -d ''; done
```

Example:

```
# curl -X POST http://localhost:8080/function/func_nodeinfo -d ''

Hostname: 9b077a81a489

Platform: linux
Arch: arm
CPU count: 1
Uptime: 776839
```

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

> Why not start the code on play-with-docker.com and then configure a Github repository to send webhooks to the API Gateway?
