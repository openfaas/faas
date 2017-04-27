// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricOptions to be used by web handlers
type MetricOptions struct {
	GatewayFunctionInvocation *prometheus.CounterVec
	GatewayFunctionsHistogram *prometheus.HistogramVec
	ServiceReplicasCounter    *prometheus.GaugeVec
}

// PrometheusHandler Bootstraps prometheus for metrics collection
func PrometheusHandler() http.Handler {
	return prometheus.Handler()
}

// BuildMetricsOptions builds metrics for tracking functions in the API gateway
func BuildMetricsOptions() MetricOptions {
	gatewayFunctionsHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "gateway_functions_seconds",
		Help: "Function time taken",
	}, []string{"function_name"})

	gatewayFunctionInvocation := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_function_invocation_total",
			Help: "Individual function metrics",
		},
		[]string{"function_name", "code"},
	)

	serviceReplicas := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
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
func RegisterMetrics(metricsOptions MetricOptions) {
	prometheus.Register(metricsOptions.GatewayFunctionInvocation)
	prometheus.Register(metricsOptions.GatewayFunctionsHistogram)
	prometheus.Register(metricsOptions.ServiceReplicasCounter)
}
