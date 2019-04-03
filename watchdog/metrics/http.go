package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Http struct {
	RequestsTotal            *prometheus.CounterVec
	RequestDurationHistogram *prometheus.HistogramVec
}

func NewHttp() Http {
	return Http{
		RequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "total HTTP requests processed",
		}, []string{"code", "method"}),
		RequestDurationHistogram: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Seconds spent serving HTTP requests.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"code", "method"}),
	}
}
