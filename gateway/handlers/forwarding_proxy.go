package handlers

import (
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/types"
)

// MakeForwardingProxyHandler create a handler which forwards HTTP requests
func MakeForwardingProxyHandler(proxy *httputil.ReverseProxy, log *logrus.Logger, metrics metrics.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.String()

		log.Printf("> Forwarding [%s] to %s", r.Method, r.URL.String())
		start := time.Now()

		writeAdapter := types.NewWriteAdapter(w)
		proxy.ServeHTTP(writeAdapter, r)

		d := time.Since(start)
		seconds := d.Seconds()
		log.Printf("< [%s] - %d took %f seconds\n", r.URL.String(),
			writeAdapter.GetHeaderCode(), seconds)

		forward := "/function/"
		if startsWith(uri, forward) {
			log.Printf("function=%s", uri[len(forward):])

			service := uri[len(forward):]

			metrics.GatewayFunctionsHistogram(
				map[string]string{"function_name": service},
				d,
			)

			code := strconv.Itoa(writeAdapter.GetHeaderCode())
			metrics.GatewayFunctionInvocation(map[string]string{
				"function_name": service,
				"code":          code,
			})
		}
	}
}

func startsWith(value, token string) bool {
	return len(value) > len(token) && strings.Index(value, token) == 0
}
