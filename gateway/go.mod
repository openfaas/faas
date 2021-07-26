module github.com/openfaas/faas/gateway

go 1.15

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/nats-io/nats-server/v2 v2.3.2 // indirect
	github.com/nats-io/nats-streaming-server v0.22.0 // indirect
	github.com/openfaas/faas-provider v0.17.3
	github.com/openfaas/nats-queue-worker v0.0.0-20200512211843-8e9eefd5a320
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	go.uber.org/goleak v1.1.10
)
