TAG?=latest

.PHONY: build
build:
	./build.sh

.PHONY: build-gateway
build-gateway:
	(cd gateway; ./build.sh latest-dev)

.PHONY: test-ci
test-ci:
	./contrib/ci.sh

.PHONY: ci-armhf-build
ci-armhf-build:
	(cd gateway; ./build.sh $(TAG))

.PHONY: ci-armhf-push
ci-armhf-push:
	(cd gateway; ./push.sh $(TAG))

.PHONY: ci-arm64-build
ci-arm64-build:
	(cd gateway; ./build.sh $(TAG))

.PHONY: ci-arm64-push
ci-arm64-push:
	(cd gateway; ./push.sh $(TAG))
