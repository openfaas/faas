.PHONY: build

build:
	./build.sh
build-gateway:
	(cd gateway; ./build.sh latest-dev)
test-unit:
	go test $(shell go list ./... | grep -v /vendor/) -cover