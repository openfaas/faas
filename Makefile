.PHONY: build

build:
	./build.sh
build-gateway:
	(cd gateway; ./build.sh latest-dev)
