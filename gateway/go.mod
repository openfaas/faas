module github.com/openfaas/faas/gateway

go 1.26.0

require (
	github.com/docker/distribution v2.8.3+incompatible
	github.com/gorilla/mux v1.8.1
	github.com/openfaas/faas-provider v0.25.12
	github.com/openfaas/nats-queue-worker v0.0.0-20250415083406-ed5414157b54
	github.com/prometheus/client_golang v1.25.0
	github.com/prometheus/client_model v0.6.3
	go.uber.org/goleak v1.3.0
	golang.org/x/sync v0.24.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-msgpack/v2 v2.1.3 // indirect
	github.com/hashicorp/raft v1.7.3 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/minio/highwayhash v1.0.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/jwt/v2 v2.8.0 // indirect
	github.com/nats-io/nats.go v1.54.0 // indirect
	github.com/nats-io/nkeys v0.4.16 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nats-io/stan.go v0.10.4 // indirect
	github.com/prometheus/common v0.72.1 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	go.etcd.io/bbolt v1.4.3 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

// replace github.com/openfaas/faas-provider => ../../faas-provider
