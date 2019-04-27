.PHONY: all
all: ci-armhf-build ci-armhf-push

.PHONY: ci-armhf-build
ci-armhf-build:
	docker build -t functions/alertmanager:0.16.1-armhf .

.PHONY: ci-armhf-push
ci-armhf-push:
	docker push functions/alertmanager:0.16.1-armhf
