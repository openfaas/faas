export DOCKER_CLI_EXPERIMENTAL=enabled

PLATFORM := "linux/amd64,linux/arm/v7,linux/arm64"

TAG?=latest
SERVER?=docker.io
OWNER?=alexellis2
NAME=gateway

.PHONY: build-local
build-local:
	@echo $(SERVER)/$(OWNER)/$(NAME):$(TAG) \
	&& docker buildx create --use --name=multiarch --node multiarch \
	&& docker buildx build \
		--progress=plain \
		--platform linux/amd64 \
		--output "type=docker,push=false" \
		--tag $(SERVER)/$(OWNER)/$(NAME):$(TAG) .

.PHONY: build-push
build-push:
	@echo $(SERVER)/$(OWNER)/$(NAME):$(TAG) \
	&& docker buildx create --use --name=multiarch --node multiarch \
	&& docker buildx build \
		--progress=plain \
		--platform $(PLATFORM) \
		--output "type=image,push=true" \
		--tag $(SERVER)/$(OWNER)/$(NAME):$(TAG) .
