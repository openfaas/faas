## Getting started with NodeJS on OpenFaaS

This guide walks you through creating your first Node.js function.

#### Install Dependencies

We're using "Docker Swarm" with Docker 17.05 or later.

###### faas-cli

Let's start by installing the CLI (command-line interface) for the
project, named [faas-cli](https://github.com/alexellis/faas-cli).
The CLI simplifies building and deploying functions.

```
$ curl -sSL https://cli.openfaas.com | sudo sh
```

#### Initialize the OpenFaaS Service

We have the dependencies, now launch OpenFaaS using Docker Swarm
by following these
[instructions](https://github.com/alexellis/faas/blob/master/guide/deployment_swarm.md).

#### Build the function

Build out a simple function directory...
```
$ mkdir -p ./functions/hello-node
$ cd ./functions
```

...and create a `handler.js` file in the `./hello-node` directory.
This file will house the function.
###### ./hello-node/handler.js
```
"use strict"

module.exports = (req) => console.log('Hello! You said:', req)
```

You can configure your OpenFaaS functions with a YAML file. The CLI will
use this for building and deploying your function.
###### hello-node.yml

```
provider:
  name: faas
  gateway: http://localhost:8080

functions:
  hello-node:
    lang: node
    handler: ./hello-node
    image: hello-node
```

It's also possible to replicate the above using the CLI. The
`handler.js` file will contain a boilerplate function, just change it to
match the above handler.
```
$ mkdir ./functions
$ cd functions
$ faas-cli new --lang node hello-node
```

The directory structure should now look like this:
```
functions
|--hello-node.yml
|--hello-node
   |--handler.js
```

And now it's build time!
```
$ faas-cli build -f ./hello-node.yml
```

...deploy to your local OpenFaaS instance.
```
$ faas-cli deploy -f ./hello-node.yml
```

You should be all set to test out your function. You can do this on the
command line with the `faas-cli invoke` command:
```
$ echo hurray! | faas-cli invoke -f ./hello-node.yml --name hello-node
```

Listing all running functions can be done with the `list` command.
```
$ faas-cli list
```

Or view running functions from a specific `.yml` file
```
$ faas-cli list -f ./hello-node.yml
```

To remove your function...
```
faas-cli remove hello-node
```

Adding or invoking functions can also be done by visiting the UI in
browser at `localhost:8080`

#### Remote hosts / Multi-Node Clusters

It's possible to develop locally but deploy the function remotely. If
the OpenFaaS service is running remotely or running as part of a multi-node
Docker Swarm, a few additional steps must be taken.

Add a remote URL and a repository tag to the image in `./hello-node.yml`
```
provider:
  gateway: <ip-address>:8080
...
functions:
  hello-node:
    image: aafrey/hello-node
...
```

Build and push the image to a repository and deploy.
```
$ faas-cli build -f ./hello-node.yml
$ faas-cli push -f ./hello-node.yml
$ faas-cli deploy -f ./hello-node.yml
```

#### Function Dependencies

What about including modules? Just like with any NodeJS project,
dependencies can be included in a `package.json` file and they will
be bundled up with the function.

In our example, the directory structure would look like this:
```
functions
|--hello-node.yml
|--hello_node
   |--handler.js
   |--package.json
```

#### That's a wrap!

That's really all there is to setting up your own serverless/FaaS
infrastructure and deploying a NodeJS function to it! As a follow up
exercise spend some time doing the following:
* Browse through the `./functions` directory. There should be some
  additional directories that were created during the `build` phase.
  Looking through some of the new files will give insight in to how the
  functions are bundled, also how you might tweak some of these files to
  change behavior.
* There are also numerous resources listed in the official [OpenFaaS](https://github.com/alexellis.com/faas) repo on GitHub, from documentation, to blog posts, to sample functions so check it out!
* For debugging live functions, there is some discussion in issue
  [#223](https://github.com/openfaas/faas/issues/223)

> Guide provided by Austin Frey
