package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricOptions to be used by web handlers
type MetricOptions struct {
	GatewayRequestsTotal         prometheus.Counter
	GatewayServerlessServedTotal prometheus.Counter
	GatewayFunctions             prometheus.Histogram
	GatewayFunctionInvocation    *prometheus.CounterVec
}

// PrometheusHandler Bootstraps prometheus for metrics collection
func PrometheusHandler() http.Handler {
	return prometheus.Handler()
}

func BuildMetricsOptions() MetricOptions {
	GatewayRequestsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_requests_total",
		Help: "Total amount of HTTP requests to the gateway",
	})
	GatewayServerlessServedTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_serverless_invocation_total",
		Help: "Total amount of serverless function invocations",
	})
	GatewayFunctions := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "gateway_functions",
		Help: "Gateway functions",
	})
	GatewayFunctionInvocation := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_function_invocation_total",
			Help: "Individual function metrics",
		},
		[]string{"function_name"},
	)

	metricsOptions := MetricOptions{
		GatewayRequestsTotal:         GatewayRequestsTotal,
		GatewayServerlessServedTotal: GatewayServerlessServedTotal,
		GatewayFunctions:             GatewayFunctions,
		GatewayFunctionInvocation:    GatewayFunctionInvocation,
	}

	return metricsOptions
}

func RegisterMetrics(metricsOptions MetricOptions) {
	prometheus.Register(metricsOptions.GatewayRequestsTotal)
	prometheus.Register(metricsOptions.GatewayServerlessServedTotal)
	prometheus.Register(metricsOptions.GatewayFunctions)
	prometheus.Register(metricsOptions.GatewayFunctionInvocation)
}
