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

	return &metrics.VectorQueryResponse{}, nil
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
	if len(rr.Body.String()) == 0 {
		t.Errorf("Want content-length > 0, got: %d", len(rr.Body.String()))
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
				Name:     "echo",
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
