[![build status](https://circleci.com/gh/docker/cli.svg?style=shield)](https://circleci.com/gh/docker/cli/tree/master)

docker/cli
==========

This repository is the home of the cli used in the Docker CE and
Docker EE products.

It's composed of 3 main folders

* `/cli` - all the commands code.
* `/cmd/docker` - the entrypoint of the cli, aka the main.

Development
===========

### Build locally

```
$ make build
```

```
$ make clean
```

You will need [gox](https://github.com/mitchellh/gox) for this one:

```
$ make cross
```

If you don't have [gox](https://github.com/mitchellh/gox), you can use the "in-container" version of `make cross`, listed below.

### Build inside container

```
$ make -f docker.Makefile build
```

```
$ make -f docker.Makefile clean
```

```
$ make -f docker.Makefile cross
```

### In-container development environment

```
$ make -f docker.Makefile dev
```

Then you can use the [build locally](#build-locally) commands:

```
$ make build
```

```
$ make clean
```

Legal
=====
*Brought to you courtesy of our legal counsel. For more context,
please see the [NOTICE](https://github.com/docker/cli/blob/master/NOTICE) document in this repo.*

Use and transfer of Docker may be subject to certain restrictions by the
United States and other governments.

It is your responsibility to ensure that your use and/or transfer does not
violate applicable laws.

For more information, please see https://www.bis.doc.gov

Licensing
=========
docker/cli is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/docker/docker/blob/master/LICENSE) for the full
license text.
