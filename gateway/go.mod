module github.com/faas/gateway

go 1.15

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/gorilla/mux v1.7.3
	github.com/nats-io/nats-server/v2 v2.1.8 // indirect
	github.com/nats-io/nats-streaming-server v0.18.0 // indirect
	github.com/openfaas/faas v0.0.0-20201020105606-15b90db952d3
	github.com/openfaas/faas-provider v0.0.0-20191005090653-478f741b64cb
	github.com/openfaas/nats-queue-worker v0.0.0-20200422114215-1f4e16e1f7af
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4
	github.com/prometheus/common v0.7.0 // indirect
	go.uber.org/goleak v0.10.0
	golang.org/x/tools v0.0.0-20201021171030-d105bfabbdbe // indirect
)
