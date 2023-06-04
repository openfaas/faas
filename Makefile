TAG?=latest
NS?=openfaas

.PHONY: build-gateway
build-gateway:
	(cd gateway;  docker buildx build --platform linux/amd64 -t ${NS}/gateway:latest-dev .)


# generate Go models from the OpenAPI spec using https://github.com/contiamo/openapi-generator-go
generate:
	rm gateway/models/model_*.go || true
	openapi-generator-go generate models -s api-docs/spec.openapi.yml -o gateway/models --package-name models

# .PHONY: test-ci
# test-ci:
# 	./contrib/ci.sh
