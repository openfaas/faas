package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	types "github.com/openfaas/faas-provider/types"
)

// AddMetricsHandler wraps a http.HandlerFunc with Prometheus metrics
func AddMetricsHandler(handler http.HandlerFunc, prometheusQuery PrometheusQueryFetcher) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, r)
		upstreamCall := recorder.Result()

		if upstreamCall.Body == nil {
			log.Println("Upstream call had empty body.")
			return
		}

		defer upstreamCall.Body.Close()
		upstreamBody, _ := ioutil.ReadAll(upstreamCall.Body)

		if recorder.Code != http.StatusOK {
			log.Printf("List functions responded with code %d, body: %s",
				recorder.Code,
				string(upstreamBody))

			http.Error(w, "Metrics handler: unexpected status code from provider listing functions", http.StatusInternalServerError)
			return
		}

		var functions []types.FunctionStatus

		err := json.Unmarshal(upstreamBody, &functions)

		if err != nil {
			log.Printf("Metrics upstream error: %s, value: %s", err, string(upstreamBody))

			http.Error(w, "Unable to parse list of functions from provider", http.StatusInternalServerError)
			return
		}

		// Ensure values are empty first.
		for i := range functions {
			functions[i].InvocationCount = 0
		}

		if len(functions) > 0 {

			ns := functions[0].Namespace
			q := fmt.Sprintf(`sum(gateway_function_invocation_total{function_name=~".*.%s"}) by (function_name)`, ns)
			// Restrict query results to only function names matching namespace suffix.

			results, err := prometheusQuery.Fetch(url.QueryEscape(q))
			if err != nil {
				log.Printf("Error querying Prometheus: %s\n", err.Error())
				return
			}
			mixIn(&functions, results)
		}

		bytesOut, err := json.Marshal(functions)
		if err != nil {
			log.Printf("Error serializing functions: %s", err)
			http.Error(w, "Error writing response after adding metrics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytesOut)
	}
}

func mixIn(functions *[]types.FunctionStatus, metrics *VectorQueryResponse) {

	if functions == nil {
		return
	}

	for i, function := range *functions {
		for _, v := range metrics.Data.Result {

			if v.Metric.FunctionName == fmt.Sprintf("%s.%s", function.Name, function.Namespace) {
				metricValue := v.Value[1]
				switch value := metricValue.(type) {
				case string:
					f, err := strconv.ParseFloat(value, 64)
					if err != nil {
						log.Printf("add_metrics: unable to convert value %q for metric: %s", value, err)
						continue
					}
					(*functions)[i].InvocationCount += f
				}
			}
		}
	}
}
