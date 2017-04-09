## Base function examples

Examples of base functions are provided here.

Each one will read the request from the watchdog then print it back resulting in an HTTP 200.

| Language               | Docker image                            | Notes                                  |
|------------------------|-----------------------------------------|----------------------------------------|
| Node.js                | functions/base:node-6.9.1-alpine        | Node.js built on Alpine Linux          |
| Golang                 | functions/base:golang-1.7.5-alpine      | Golang compiled on Alpine Linux        |
| Python                 | functions/base:python-2.7-alpine        | Python 2.7 built on Alpine Linux       |
| Java                   | functions/base:openjdk-8u121-jdk-alpine | OpenJDK built on Alpine Linux |
| Busybox / shell        | functions/alpine:latest            | Busybox contains useful binaries which can be turned into a FaaS function such as `sha512sum` or `cat` |
