ifndef VERBOSE
.SILENT:
endif

# set default shell
SHELL=/bin/bash -o pipefail -o errexit

DIR:=$(strip $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))))
INIT_BUILDX=$(DIR)/../../contrib/init-buildx.sh

TAG ?= latest
NS  ?= openfaas

PLATFORMS?=linux/amd64,linux/arm,linux/arm64
OUTPUT=
PROGRESS=plain

all: build

build: ensure-buildx
	docker buildx build \
	--platform=${PLATFORMS} $(OUTPUT) \
	--progress=$(PROGRESS) \
	--pull \
	--build-arg https_proxy=$(https_proxy) --build-arg http_proxy=$(http_proxy) \
	-t $(NS)/basic-auth-plugin:$(TAG) .

# push the cross built image
push: OUTPUT=--push
push: build

# enable buildx
ensure-buildx:
	@$(INIT_BUILDX)

.PHONY: all build push ensure-buildx
