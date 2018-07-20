.PHONY: build build-gateway test-ci

build:
	./build.sh
build-gateway:
	(cd gateway; ./build.sh latest-dev)
test-ci:
	./contrib/ci.sh
