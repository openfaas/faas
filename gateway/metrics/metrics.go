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
}

// PrometheusHandler Bootstraps prometheus for metrics collection
func PrometheusHandler() http.Handler {
	return prometheus.Handler()
}
