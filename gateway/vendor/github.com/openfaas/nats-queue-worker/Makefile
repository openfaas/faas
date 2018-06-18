TAG?=latest

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t functions/queue-worker:$(TAG) .

push:
	docker push functions/queue-worker:$(TAG)

all: build
