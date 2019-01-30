.PHONY: all
all: ci-arm64-build ci-arm64-push

.PHONY: ci-arm64-build
ci-arm64-build:
	docker build -t functions/prometheus:2.7.0-arm64 .

.PHONY: ci-arm64-push
ci-arm64-push:
	docker push functions/prometheus:2.7.0-arm64
