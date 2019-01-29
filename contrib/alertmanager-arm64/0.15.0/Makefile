.PHONY: ci-arm64-build
ci-arm64-build:
	docker build -t functions/alertmanager:0.15.0-arm64 .

.PHONY: ci-arm64-push
ci-arm64-push:
	docker push functions/alertmanager:0.15.0-arm64