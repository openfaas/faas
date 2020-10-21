ifndef VERBOSE
.SILENT:
endif

# set default shell
SHELL=/bin/bash -o pipefail -o errexit

TAG ?= latest
NS  ?= openfaas

GIT_COMMIT_MESSAGE?=$(git log -1 --pretty=%B 2>&1 | head -n 1)
GIT_COMMIT_SHA?=$(git rev-list -1 HEAD)
VERSION?=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///' || echo dev)

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
	--build-arg GIT_COMMIT_MESSAGE="$(GIT_COMMIT_MESSAGE)" \
	--build-arg GIT_COMMIT_SHA="$(GIT_COMMIT_SHA)" \
	--build-arg VERSION="$(VERSION:-dev)" \
	-t $(NS)/gateway:$(TAG) .

# push the cross built image
push: OUTPUT=--push
push: build

# enable buildx
ensure-buildx:
# this is required for cloudbuild
ifeq ("$(wildcard $(INIT_BUILDX))","")
	@./init-buildx.sh
else
	@exec $(INIT_BUILDX)
endif
	@echo "done"

.PHONY: all build push ensure-buildx
