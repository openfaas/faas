package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/pkg/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

// HTTPNotifier notify about HTTP request/response
type HTTPNotifier interface {
	Notify(method string, URL string, originalURL string, statusCode int, event string, duration time.Duration)
}

func urlToLabel(path string) string {
	if len(path) > 0 {
		path = strings.TrimRight(path, "/")
	}
	if path == "" {
		path = "/"
	}
	return path
}

// PrometheusFunctionNotifier records metrics to Prometheus
type PrometheusFunctionNotifier struct {
	Metrics *metrics.MetricOptions
	//FunctionNamespace default namespace of the function
	FunctionNamespace string
}

// Notify records metrics in Prometheus
func (p PrometheusFunctionNotifier) Notify(method string, URL string, originalURL string, statusCode int, event string, duration time.Duration) {
	serviceName := middleware.GetServiceName(originalURL)
	if len(p.FunctionNamespace) > 0 {
		if !strings.Contains(serviceName, ".") {
			serviceName = fmt.Sprintf("%s.%s", serviceName, p.FunctionNamespace)
		}
	}

	code := strconv.Itoa(statusCode)
	labels := prometheus.Labels{"function_name": serviceName, "code": code}

	if event == "completed" {
		seconds := duration.Seconds()
		p.Metrics.GatewayFunctionsHistogram.
			With(labels).
			Observe(seconds)

		p.Metrics.GatewayFunctionInvocation.
			With(labels).
			Inc()
	} else if event == "started" {
		p.Metrics.GatewayFunctionInvocationStarted.WithLabelValues(serviceName).Inc()
	}

}

// LoggingNotifier notifies a log about a request
type LoggingNotifier struct {
}

// Notify the LoggingNotifier about a request
func (LoggingNotifier) Notify(method string, URL string, originalURL string, statusCode int, event string, duration time.Duration) {
	if event == "completed" {
		log.Printf("Forwarded [%s] to %s - [%d] - %.4fs", method, originalURL, statusCode, duration.Seconds())
	}
}
