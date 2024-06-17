module github.com/openfaas/faas/gateway

go 1.22

require (
	github.com/docker/distribution v2.8.3+incompatible
	github.com/gorilla/mux v1.8.1
	github.com/openfaas/faas-provider v0.25.3
	github.com/openfaas/nats-queue-worker v0.0.0-20231219105451-b94918cb8a24
	github.com/prometheus/client_golang v1.19.1
	github.com/prometheus/client_model v0.6.1
	go.uber.org/goleak v1.3.0
	golang.org/x/sync v0.7.0
)

// replace github.com/openfaas/faas-provider => ../../faas-provider

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/nats-io/nats.go v1.36.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nats-io/stan.go v0.10.4 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/prometheus/common v0.54.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)
