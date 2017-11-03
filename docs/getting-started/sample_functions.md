## Sample functions

We have packaged some simple starter functions in the Docker stack when deployed on Swarm, so as soon as you open the OpenFaaS UI you will see them listed down the left-hand side.

Here are a few of the functions:

* Echo function (echoit) - echos any received text back to the caller (wraps Linux `cat` binary)
* Markdown to HTML renderer (markdownrender) - takes .MD input and produces HTML (Golang)
* Docker Hub Stats function (hubstats) - queries the count of images for a user on the Docker Hub (Golang)
* Node Info (nodeinfo) function - gives you the OS architecture and detailled info about the CPUS (Node.js)
* Webhook stasher function (webhookstash) - saves webhook body into container's filesystem - even binaries (Golang)



<!-- ## Learn the CLI -->

You can now grab a coffee and start learning how to create your first function with the CLI:

[Morning coffee with the OpenFaaS CLI](https://blog.alexellis.io/quickstart-openfaas-cli/)
