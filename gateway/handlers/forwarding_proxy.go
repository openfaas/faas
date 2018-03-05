package handlers

import (
	"context"
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
		serviceName := getServiceName(requestURL)

		log.Printf("> Forwarding [%s] to %s", r.Method, requestURL)

		start := time.Now()

		statusCode, err := forwardRequest(w, r, proxy.Client, baseURL, requestURL, proxy.Timeout)

		if err != nil {
			log.Printf("error with upstream request to: %s, %s\n", requestURL, err.Error())
		}

		seconds := time.Since(start).Seconds()
		log.Printf("< [%s] - %d took %f seconds\n", r.URL.String(),
			statusCode, seconds)

		if len(serviceName) > 0 {
			metrics.GatewayFunctionsHistogram.
				WithLabelValues(serviceName).
				Observe(seconds)

			code := strconv.Itoa(statusCode)

			metrics.GatewayFunctionInvocation.
				With(prometheus.Labels{"function_name": serviceName, "code": code}).
				Inc()
		}

	}
}

func forwardRequest(w http.ResponseWriter, r *http.Request, proxyClient *http.Client, baseURL string, requestURL string, timeout time.Duration) (int, error) {

	upstreamReq, _ := http.NewRequest(r.Method, baseURL+requestURL, nil)
	copyHeaders(upstreamReq.Header, &r.Header)

	upstreamReq.Header["X-Forwarded-For"] = []string{r.RemoteAddr}

	if r.Body != nil {
		defer r.Body.Close()
		upstreamReq.Body = r.Body
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, resErr := proxyClient.Do(upstreamReq.WithContext(ctx))
	if resErr != nil {
		badStatus := http.StatusBadGateway
		w.WriteHeader(badStatus)
		return badStatus, resErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	copyHeaders(w.Header(), &res.Header)

	// Write status code
	w.WriteHeader(res.StatusCode)

	if res.Body != nil {
		// Copy the body over
		io.CopyBuffer(w, res.Body, nil)
	}

	return res.StatusCode, nil
}

func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		(destination)[k] = vClone
	}
}

func getServiceName(urlValue string) string {
	var serviceName string
	forward := "/function/"
	if startsWith(urlValue, forward) {
		serviceName = urlValue[len(forward):]
	}
	return serviceName
}

func startsWith(value, token string) bool {
	return len(value) > len(token) && strings.Index(value, token) == 0
}
