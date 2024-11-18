// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Copyright (c) Alex Ellis 2017. All rights reserved.

package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricOptions to be used by web handlers
type MetricOptions struct {
	GatewayFunctionInvocation        *prometheus.CounterVec
	GatewayFunctionsHistogram        *prometheus.HistogramVec
	GatewayFunctionInvocationStarted *prometheus.CounterVec

	ServiceReplicasGauge *prometheus.GaugeVec
}

// ServiceMetricOptions provides RED metrics
type ServiceMetricOptions struct {
	Histogram *prometheus.HistogramVec
	Counter   *prometheus.CounterVec
}

// Synchronize to make sure MustRegister only called once
var once = sync.Once{}

// RegisterExporter registers with Prometheus for tracking
func RegisterExporter(exporter *Exporter) {
	once.Do(func() {
		prometheus.MustRegister(exporter)
	})
}

// PrometheusHandler Bootstraps prometheus for metrics collection
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// BuildMetricsOptions builds metrics for tracking functions in the API gateway
func BuildMetricsOptions() MetricOptions {
	gatewayFunctionsHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "gateway_functions_seconds",
		Help: "Function time taken",
	}, []string{"function_name", "code"})

	gatewayFunctionInvocation := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gateway",
			Subsystem: "function",
			Name:      "invocation_total",
			Help:      "Function metrics",
		},
		[]string{"function_name", "code"},
	)

	serviceReplicas := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gateway",
			Name:      "service_count",
			Help:      "Current count of replicas for function",
		},
		[]string{"function_name"},
	)

	gatewayFunctionInvocationStarted := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gateway",
			Subsystem: "function",
			Name:      "invocation_started",
			Help:      "The total number of function HTTP requests started.",
		},
		[]string{"function_name"},
	)

	metricsOptions := MetricOptions{
		GatewayFunctionsHistogram:        gatewayFunctionsHistogram,
		GatewayFunctionInvocation:        gatewayFunctionInvocation,
		ServiceReplicasGauge:             serviceReplicas,
		GatewayFunctionInvocationStarted: gatewayFunctionInvocationStarted,
	}

	return metricsOptions
}
