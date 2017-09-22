package handlers

import (
	"net/http"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/docker/docker/client"
)

// MakeUpdateFunctionHandler request to update an existing function with new configuration such as image, parameters etc.
func MakeUpdateFunctionHandler(metricsOptions metrics.MetricOptions, c *client.Client, maxRestarts uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.WriteHeader(http.StatusNotImplemented)
	}
}
