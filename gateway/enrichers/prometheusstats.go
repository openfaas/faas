package enrichers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/requests"
)

// AddPrometheusMetrics TODO
func PrometheusMetrics(host string, port int) EnricherFunc {
	return func(h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			h(&PrometheusStatsResponseWriter{
				rw:   w,
				host: host,
				port: port,
			}, r)
		}
	}
}

// PrometheusStatsResponseWriter TODO
type PrometheusStatsResponseWriter struct {
	rw            http.ResponseWriter
	headerWritten bool
	host          string
	port          int
}

// Header TODO
func (prw *PrometheusStatsResponseWriter) Header() http.Header {
	return prw.rw.Header()
}

// WriteHeader TODO
func (prw *PrometheusStatsResponseWriter) WriteHeader(status int) {
	prw.rw.WriteHeader(status)
}

// Write TODO
func (prw *PrometheusStatsResponseWriter) Write(p []byte) (int, error) {
	if !prw.headerWritten {
		prw.headerWritten = true
	}

	var functions []requests.Function

	err := json.Unmarshal(p, &functions)
	if err != nil {
		log.Println(err)

		prw.WriteHeader(http.StatusInternalServerError)
		return prw.Write([]byte("Error parsing metrics from upstream provider/backend."))
	}

	client := http.Client{}

	prometheusQuery := metrics.NewPrometheusQuery(prw.host, prw.port, &client)

	expr := "sum(gateway_function_invocation_total%7Bfunction_name%3D~%22.*%22%2C+code%3D~%22.*%22%7D)+by+(function_name%2C+code)"
	results, fetchErr := prometheusQuery.Fetch(expr)
	if fetchErr != nil {
		log.Printf("Error querying Prometheus API: %s\n", fetchErr.Error())

		prw.WriteHeader(http.StatusOK)
		return prw.Write(p)
	}

	mixIn(&functions, results)

	bytesOut, marshalErr := json.Marshal(functions)
	if marshalErr != nil {
		log.Println(marshalErr)
		prw.WriteHeader(http.StatusInternalServerError)
		return prw.Write([]byte("Error parsing metrics from upstream provider/backend."))
	}

	return prw.rw.Write(bytesOut)
}

func mixIn(functions *[]requests.Function, metrics *metrics.VectorQueryResponse) {
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
