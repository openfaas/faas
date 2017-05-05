BaseFunction for Golang
=========================

You will find a Dockerfile for Linux and one for Windows so that you can run serverless functions in a mixed-OS swarm.

Dockerfile for Windows
* [Dockerfile.win](https://github.com/alexellis/faas/blob/master/sample-functions/BaseFunctions/golang/Dockerfile.win)

This function reads STDIN then prints it back - you can use it to create an "echo service" or as a basis to write your own function.

