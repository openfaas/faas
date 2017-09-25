package handlers

import (
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/types"
	"github.com/prometheus/client_golang/prometheus"
)

// MakeForwardingProxyHandler create a handler which forwards HTTP requests
func MakeForwardingProxyHandler(proxy *httputil.ReverseProxy, metrics *metrics.MetricOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.String()

		log.Printf("> Forwarding [%s] to %s", r.Method, r.URL.String())
		start := time.Now()

		writeAdapter := types.NewWriteAdapter(w)
		proxy.ServeHTTP(writeAdapter, r)

		seconds := time.Since(start).Seconds()
		log.Printf("< [%s] - %d took %f seconds\n", r.URL.String(), writeAdapter.GetHeaderCode(), seconds)

		forward := "/function/"
		if startsWith(uri, forward) {
			log.Printf("function=%s", uri[len(forward):])

			service := uri[len(forward):]

			metrics.GatewayFunctionsHistogram.
				WithLabelValues(service).
				Observe(seconds)

			code := strconv.Itoa(writeAdapter.GetHeaderCode())

			metrics.GatewayFunctionInvocation.With(prometheus.Labels{"function_name": service, "code": code}).Inc()
		}
	}
}

func startsWith(value, token string) bool {
	return len(value) > len(token) && strings.Index(value, token) == 0
}
