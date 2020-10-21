TAG?=latest

build: build-auth-basic build-gateway

build-gateway:
	(TAG=latest-dev make -C gateway push)

build-auth-basic:
	(TAG=latest-dev make -C auth/basic-auth push)

test-ci:
	./contrib/ci.sh

.PHONY: build build-auth-basic build-gateway test-ci