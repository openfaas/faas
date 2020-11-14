TAG?=latest
NS?=openfaas

.PHONY: build-gateway
build-gateway:
	(cd gateway;  docker buildx build --platform linux/amd64 -t $NS/gateway:latest-dev .)

.PHONY: test-ci
test-ci:
	./contrib/ci.sh
