module github.com/openfaas/faas/gateway

go 1.16

require (
	github.com/docker/distribution v2.8.1+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/openfaas/faas-provider v0.18.7
	github.com/openfaas/nats-queue-worker v0.0.0-20210726161954-ada9a31504c9
	github.com/prometheus/client_golang v1.11.1
	github.com/prometheus/client_model v0.2.0
	go.uber.org/goleak v1.1.10
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
)
