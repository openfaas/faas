module github.com/openfaas/faas/gateway

go 1.15

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/nats-io/jwt v1.2.2 // indirect
	github.com/nats-io/nats-streaming-server v0.20.0 // indirect
	github.com/nats-io/stan.go v0.8.2 // indirect
	github.com/openfaas/faas-provider v0.16.2
	github.com/openfaas/nats-queue-worker v0.0.0-20200512211843-8e9eefd5a320
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/procfs v0.3.0 // indirect
	go.uber.org/goleak v1.1.10
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
)
