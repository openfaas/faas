TAG?=latest
NS?=openfaas

COMMIT ?= $(shell git rev-parse HEAD)

.PHONY: build-gateway
build-gateway:
	(cd gateway;  docker buildx build --platform linux/amd64 --load -t  ${NS}/gateway:${COMMIT} -t ${NS}/gateway:latest-dev .)


kind-load:
	kind --name of-tracing load docker-image ${NS}/gateway:${COMMIT}
# .PHONY: test-ci
# test-ci:
# 	./contrib/ci.sh
