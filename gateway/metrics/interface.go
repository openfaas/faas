package metrics

import "time"

//go:generate moq -out metrics_mocks.go . Metrics

// Metrics defines an interface for metric such as Prometheus should
// implement to send metric to the server
type Metrics interface {
	GatewayFunctionInvocation(labels map[string]string)
	GatewayFunctionsHistogram(labels map[string]string, duration time.Duration)
	ServiceReplicasCounter(labels map[string]string, replicas float64)
}
