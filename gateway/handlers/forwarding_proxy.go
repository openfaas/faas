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

type HTTPNotifier interface {
	Notify(method string, URL string, statusCode int, duration time.Duration)
}

// MakeForwardingProxyHandler create a handler which forwards HTTP requests
func MakeForwardingProxyHandler(proxy *types.HTTPClientReverseProxy, notifiers []HTTPNotifier) http.HandlerFunc {
	baseURL := proxy.BaseURL.String()
	if strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL[0 : len(baseURL)-1]
	}

	return func(w http.ResponseWriter, r *http.Request) {

		requestURL := r.URL.String()

		start := time.Now()

		statusCode, err := forwardRequest(w, r, proxy.Client, baseURL, requestURL, proxy.Timeout)

		seconds := time.Since(start)
		if err != nil {
			log.Printf("error with upstream request to: %s, %s\n", requestURL, err.Error())
		}
		for _, notifier := range notifiers {
			notifier.Notify(r.Method, requestURL, statusCode, seconds)
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

type PrometheusFunctionNotifier struct {
	Metrics *metrics.MetricOptions
}

func (p PrometheusFunctionNotifier) Notify(method string, URL string, statusCode int, duration time.Duration) {
	seconds := duration.Seconds()
	serviceName := getServiceName(URL)

	p.Metrics.GatewayFunctionsHistogram.
		WithLabelValues(serviceName).
		Observe(seconds)

	code := strconv.Itoa(statusCode)

	p.Metrics.GatewayFunctionInvocation.
		With(prometheus.Labels{"function_name": serviceName, "code": code}).
		Inc()
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

type LoggingNotifier struct {
}

func (LoggingNotifier) Notify(method string, URL string, statusCode int, duration time.Duration) {
	log.Printf("Forwarded [%s] to %s - [%d] - %f seconds", method, URL, statusCode, duration.Seconds())
}
