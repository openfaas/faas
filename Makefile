.PHONY: build build-gateway test-ci
TAG?=latest

build:
	./build.sh
build-gateway:
	(cd gateway; ./build.sh latest-dev)
test-ci:
	./contrib/ci.sh
ci-armhf:
	(cd gateway; ./build.sh $(TAG)-armhf)
