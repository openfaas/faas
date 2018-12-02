.PHONY: build build-gateway test-ci ci-armhf-build ci-armhf-push ci-arm64-build ci-arm64-push
TAG?=latest

build:
	./build.sh

build-gateway:
	(cd gateway; ./build.sh latest-dev)

test-ci:
	./contrib/ci.sh

ci-armhf-build:
	(cd gateway; ./build.sh $(TAG))

ci-armhf-push:
	(cd gateway; ./push.sh $(TAG))

ci-arm64-build:
	(cd gateway; ./build.sh $(TAG))

ci-arm64-push:
	(cd gateway; ./push.sh $(TAG))
