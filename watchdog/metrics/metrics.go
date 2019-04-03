package metrics

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer provides instrumentation for HTTP calls
type MetricsServer struct {
	s    *http.Server
	port int
}

// Register binds a HTTP server to expose Prometheus metrics
func (m *MetricsServer) Register(metricsPort int) {

	m.port = metricsPort

	readTimeout := time.Millisecond * 500
	writeTimeout := time.Millisecond * 500

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	m.s = &http.Server{
		Addr:           fmt.Sprintf(":%d", metricsPort),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
		Handler:        metricsMux,
	}

}

// Serve http traffic in go routine, non-blocking
func (m *MetricsServer) Serve(cancel chan bool) {
	log.Printf("Metrics server. Port: %d\n", m.port)

	go func() {
		if err := m.s.ListenAndServe(); err != http.ErrServerClosed {
			panic(fmt.Sprintf("metrics error ListenAndServe: %v\n", err))
		}
	}()

	go func() {
		select {
		case <-cancel:
			log.Printf("metrics server shutdown\n")

			m.s.Shutdown(context.Background())
		}
	}()
}

// InstrumentHandler returns a handler which records HTTP requests
// as they are made
func InstrumentHandler(next http.HandlerFunc, _http Http) http.HandlerFunc {
	return promhttp.InstrumentHandlerCounter(_http.RequestsTotal,
		promhttp.InstrumentHandlerDuration(_http.RequestDurationHistogram, next))
}
