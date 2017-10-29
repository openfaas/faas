package tests

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/requests"
)

type FakePrometheusQueryFetcher struct {
}

func (q FakePrometheusQueryFetcher) Fetch(query string) (*metrics.VectorQueryResponse, error) {
	val := []byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"code":"200","function_name":"func_echoit"},"value":[1509267827.752,"1"]}]}}`)
	queryRes := metrics.VectorQueryResponse{}
	err := json.Unmarshal(val, &queryRes)
	return &queryRes, err
}

func makeFakePrometheusQueryFetcher() FakePrometheusQueryFetcher {
	return FakePrometheusQueryFetcher{}
}

func Test_PrometheusMetrics_MixedInto_Services(t *testing.T) {
	functionsHandler := makeFunctionsHandler()
	fakeQuery := makeFakePrometheusQueryFetcher()

	handler := metrics.AddMetricsHandler(functionsHandler, fakeQuery)

	rr := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/system/functions", nil)
	handler.ServeHTTP(rr, request)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Want application/json content-type, got: %s", rr.Header().Get("Content-Type"))
	}
	body := rr.Body.String()
	if len(body) == 0 {
		t.Errorf("Want content-length > 0, got: %d", len(rr.Body.String()))
	}
	results := []requests.Function{}
	json.Unmarshal([]byte(rr.Body.String()), &results)
	if len(results) == 0 {
		t.Errorf("Want %d function, got: %d", 1, len(results))
	}
	if results[0].InvocationCount != 1 {
		t.Errorf("InvocationCount want: %d , got: %f", 1, results[0].InvocationCount)
	}
}

func Test_FunctionsHandler_ReturnsJSONAndOneFunction(t *testing.T) {
	functionsHandler := makeFunctionsHandler()

	rr := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/system/functions", nil)
	if err != nil {
		t.Fatal(err)
	}

	functionsHandler.ServeHTTP(rr, request)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Want application/json content-type, got: %s", rr.Header().Get("Content-Type"))
	}

	if len(rr.Body.String()) == 0 {
		t.Errorf("Want content-length > 0, got: %d", len(rr.Body.String()))
	}

}

func makeFunctionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		functions := []requests.Function{
			requests.Function{
				Name:     "func_echoit",
				Replicas: 0,
			},
		}
		bytesOut, marshalErr := json.Marshal(&functions)
		if marshalErr != nil {
			log.Fatal(marshalErr.Error())
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(bytesOut)
		if err != nil {
			log.Fatal(err)
		}
	}
}
