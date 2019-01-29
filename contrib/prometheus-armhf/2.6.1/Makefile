.PHONY: ci-armhf-build
ci-armhf-build:
	docker build -t functions/prometheus:2.6.1-armhf .

.PHONY: ci-armhf-push
ci-armhf-push:
	docker push functions/prometheus:2.6.1-armhf
