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

		if recorder.Code != http.StatusOK {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error pulling metrics from provider/backend. Status code: %d", recorder.Code)))
			return
		}

		upstreamBody, _ := ioutil.ReadAll(upstreamCall.Body)
		var functions []types.FunctionStatus

		err := json.Unmarshal(upstreamBody, &functions)

		if err != nil {
			log.Printf("Metrics upstream error: %s", err)

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error parsing metrics from upstream provider/backend."))
			return
		}

		expr := url.QueryEscape(`sum(gateway_function_invocation_total{function_name=~".*", code=~".*"}) by (function_name, code)`)
		// expr := "sum(gateway_function_invocation_total%7Bfunction_name%3D~%22.*%22%2C+code%3D~%22.*%22%7D)+by+(function_name%2C+code)"
		results, fetchErr := prometheusQuery.Fetch(expr)
		if fetchErr != nil {
			log.Printf("Error querying Prometheus API: %s\n", fetchErr.Error())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(upstreamBody)
			return
		}

		mixIn(&functions, results)

		bytesOut, marshalErr := json.Marshal(functions)
		if marshalErr != nil {
			log.Println(marshalErr)
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

	// Ensure values are empty first.
	for i := range *functions {
		(*functions)[i].InvocationCount = 0
	}

	for i, function := range *functions {
		for _, v := range metrics.Data.Result {

			if v.Metric.FunctionName == function.Name {
				metricValue := v.Value[1]
				switch metricValue.(type) {
				case string:
					// log.Println("String")
					f, strconvErr := strconv.ParseFloat(metricValue.(string), 64)
					if strconvErr != nil {
						log.Printf("Unable to convert value for metric: %s\n", strconvErr)
						continue
					}
					(*functions)[i].InvocationCount += f
					break
				}
			}
		}
	}
}
