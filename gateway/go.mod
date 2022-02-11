module github.com/openfaas/faas/gateway

go 1.16

require (
	github.com/docker/distribution v2.8.0+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/openfaas/faas-provider v0.18.7
	github.com/openfaas/nats-queue-worker v0.0.0-20210726161954-ada9a31504c9
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	go.uber.org/goleak v1.1.10
)
