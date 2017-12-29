// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package prometheus

import (
	"net/http"
	"time"

	"github.com/openfaas/faas/gateway/metrics"
	promclient "github.com/prometheus/client_golang/prometheus"
)

// NewMetrics creates a new prometheus metrics writer
func NewMetrics() metrics.Metrics {
	metricsOptions := BuildMetricsOptions()
	RegisterMetrics(&metricsOptions)

	return &Client{&metricsOptions}
}

// Client defines a prometheus metrics client which implements the Metrics interface
type Client struct {
	options *MetricOptions
}

// GatewayFunctionInvocation updates the number of function calls in prometheus
func (c *Client) GatewayFunctionInvocation(labels map[string]string) {
	c.options.GatewayFunctionInvocation.With(labels).Inc()
}

// GatewayFunctionsHistogram updates the time take to invoke a function in prometheus
func (c *Client) GatewayFunctionsHistogram(labels map[string]string, duration time.Duration) {
	c.options.GatewayFunctionsHistogram.
		With(labels).
		Observe(duration.Seconds())
}

// ServiceReplicasCounter sets the number of replicas for a service
func (c *Client) ServiceReplicasCounter(labels map[string]string, replicas float64) {
	c.options.ServiceReplicasCounter.
		With(labels).
		Set(replicas)
}

// MetricOptions to be used by web handlers
type MetricOptions struct {
	GatewayFunctionInvocation *promclient.CounterVec
	GatewayFunctionsHistogram *promclient.HistogramVec
	ServiceReplicasCounter    *promclient.GaugeVec
}

// Handler Bootstraps prometheus for metrics collection
func Handler() http.Handler {
	return promclient.Handler()
}

// BuildMetricsOptions builds metrics for tracking functions in the API gateway
func BuildMetricsOptions() MetricOptions {
	gatewayFunctionsHistogram := promclient.NewHistogramVec(promclient.HistogramOpts{
		Name: "gateway_functions_seconds",
		Help: "Function time taken",
	}, []string{"function_name"})

	gatewayFunctionInvocation := promclient.NewCounterVec(
		promclient.CounterOpts{
			Name: "gateway_function_invocation_total",
			Help: "Individual function metrics",
		},
		[]string{"function_name", "code"},
	)

	serviceReplicas := promclient.NewGaugeVec(
		promclient.GaugeOpts{
			Name: "gateway_service_count",
			Help: "Docker service replicas",
		},
		[]string{"function_name"},
	)

	metricsOptions := MetricOptions{
		GatewayFunctionsHistogram: gatewayFunctionsHistogram,
		GatewayFunctionInvocation: gatewayFunctionInvocation,
		ServiceReplicasCounter:    serviceReplicas,
	}

	return metricsOptions
}

//RegisterMetrics registers with Prometheus for tracking
func RegisterMetrics(metricsOptions *MetricOptions) {
	promclient.Register(metricsOptions.GatewayFunctionInvocation)
	promclient.Register(metricsOptions.GatewayFunctionsHistogram)
	promclient.Register(metricsOptions.ServiceReplicasCounter)
}
