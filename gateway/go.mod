module github.com/openfaas/faas/gateway

go 1.21

require (
	github.com/docker/distribution v2.8.3+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/openfaas/faas-provider v0.25.2
	github.com/openfaas/nats-queue-worker v0.0.0-20231023101743-fa54e89c9db2
	github.com/prometheus/client_golang v1.17.0
	github.com/prometheus/client_model v0.5.0
	go.uber.org/goleak v1.2.1
	golang.org/x/sync v0.4.0
)

// replace github.com/openfaas/faas-provider => ../../faas-provider

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/nats-io/nats.go v1.31.0 // indirect
	github.com/nats-io/nkeys v0.4.5 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nats-io/stan.go v0.10.4 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
