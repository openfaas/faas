package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/requests"
)

// MakeAsyncReport makes a handler for asynchronous invocations to report back into.
func MakeAsyncReport(metrics metrics.MetricOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		report := requests.AsyncReport{}
		bytesOut, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(bytesOut, &report)

		trackInvocation(report.FunctionName, metrics, report.StatusCode)
	}
}
