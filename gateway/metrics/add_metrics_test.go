package metrics

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	types "github.com/openfaas/faas-provider/types"
)

type FakePrometheusQueryFetcher struct {
}

func (q FakePrometheusQueryFetcher) Fetch(query string) (*VectorQueryResponse, error) {
	val := []byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"code":"200","function_name":"func_echoit.openfaas-fn"},"value":[1509267827.752,"1"]}]}}`)
	queryRes := VectorQueryResponse{}
	err := json.Unmarshal(val, &queryRes)
	return &queryRes, err
}

func makeFakePrometheusQueryFetcher() FakePrometheusQueryFetcher {
	return FakePrometheusQueryFetcher{}
}

func Test_PrometheusMetrics_MixedInto_Services(t *testing.T) {
	functionsHandler := makeFunctionsHandler()
	fakeQuery := makeFakePrometheusQueryFetcher()

	handler := AddMetricsHandler(functionsHandler, fakeQuery)

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
	results := []types.FunctionStatus{}
	json.Unmarshal([]byte(rr.Body.String()), &results)
	if len(results) == 0 {
		t.Errorf("Want %d function, got: %d", 1, len(results))
	}
	if results[0].InvocationCount != 1 {
		t.Errorf("InvocationCount want: %d , got: %f", 1, results[0].InvocationCount)
	}
}

func Test_MetricHandler_ForwardsErrors(t *testing.T) {
	functionsHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("test error case"))
	}
	// explicitly set the query fetcher to nil because it should
	// not be called when a non-200 response is returned from the
	// functions handler, if it is called then the test will panic
	handler := AddMetricsHandler(functionsHandler, nil)

	rr := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/system/functions", nil)
	handler.ServeHTTP(rr, request)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusConflict)
	}

	if rr.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("Want 'text/plain; charset=utf-8' content-type, got: %s", rr.Header().Get("Content-Type"))
	}
	body := strings.TrimSpace(rr.Body.String())
	if body != "test error case" {
		t.Errorf("Want 'test error case', got: %q", body)
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
		functions := []types.FunctionStatus{
			{
				Name:      "func_echoit",
				Replicas:  0,
				Namespace: "openfaas-fn",
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
