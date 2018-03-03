package handlers

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/types"
	"github.com/prometheus/client_golang/prometheus"
)

// MakeForwardingProxyHandler create a handler which forwards HTTP requests
func MakeForwardingProxyHandler(proxy *types.HTTPClientReverseProxy, metrics *metrics.MetricOptions) http.HandlerFunc {
	baseURL := proxy.BaseURL.String()
	if strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL[0 : len(baseURL)-1]
	}

	return func(w http.ResponseWriter, r *http.Request) {

		requestURL := r.URL.String()

		log.Printf("> Forwarding [%s] to %s", r.Method, requestURL)
		start := time.Now()

		upstreamReq, _ := http.NewRequest(r.Method, baseURL+requestURL, nil)

		upstreamReq.Header["X-Forwarded-For"] = []string{r.RequestURI}

		if r.Body != nil {
			defer r.Body.Close()
			upstreamReq.Body = r.Body

		}

		res, resErr := proxy.Client.Do(upstreamReq)
		if resErr != nil {
			log.Printf("upstream client error: %s\n", resErr)
			return
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		// Populate any headers received
		for k, v := range res.Header {
			w.Header()[k] = v
		}

		// Write status code
		w.WriteHeader(res.StatusCode)

		// Copy the body over
		io.CopyBuffer(w, res.Body, nil)

		seconds := time.Since(start).Seconds()
		log.Printf("< [%s] - %d took %f seconds\n", r.URL.String(),
			res.StatusCode, seconds)

		forward := "/function/"
		if startsWith(requestURL, forward) {
			// log.Printf("function=%s", uri[len(forward):])

			service := requestURL[len(forward):]

			metrics.GatewayFunctionsHistogram.
				WithLabelValues(service).
				Observe(seconds)

			code := strconv.Itoa(res.StatusCode)

			metrics.GatewayFunctionInvocation.
				With(prometheus.Labels{"function_name": service, "code": code}).
				Inc()
		}

	}
}

func startsWith(value, token string) bool {
	return len(value) > len(token) && strings.Index(value, token) == 0
}
