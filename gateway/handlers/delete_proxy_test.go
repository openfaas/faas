package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/openfaas/faas/gateway/types"
	dto "github.com/prometheus/client_model/go"
)

func Test_getFunctionNameFromRequest_should_return_function_name_with_valid_request(t *testing.T) {

	dummyURL := "http://example.com"

	expectedFunctionName := "test_function"
	delReq := requests.DeleteFunctionRequest{
		FunctionName: expectedFunctionName,
	}
	b, _ := json.Marshal(delReq)
	req, _ := http.NewRequest(http.MethodPost, dummyURL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	actualFunctionName, _ := getFunctionNameFromRequest(req)

	if expectedFunctionName != actualFunctionName {
		t.Errorf("Got Function Name: %s, want %s\n", expectedFunctionName, actualFunctionName)
		t.Fail()
	}

	// check that the request body is once again readable
	reqData, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	dlr := requests.DeleteFunctionRequest{}
	json.Unmarshal(reqData, &dlr)

	if expectedFunctionName != dlr.FunctionName {
		t.Errorf("Got Function Name: %s for the second read, want %s\n", expectedFunctionName, dlr.FunctionName)
		t.Fail()
	}
}

func Test_getFunctionNameFromRequest_should_return_error_if_function_name_empty(t *testing.T) {

	dummyURL := "http://example.com"

	emptyFunctionName := ""
	delReq := requests.DeleteFunctionRequest{
		FunctionName: emptyFunctionName,
	}
	b, _ := json.Marshal(delReq)
	req, _ := http.NewRequest(http.MethodPost, dummyURL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	_, err := getFunctionNameFromRequest(req)

	if err == nil {
		t.Errorf("Wanted an error but got nil")
		t.Fail()
	}
}

func Test_getFunctionNameFromRequest_should_return_error_if_json_is_invalid(t *testing.T) {

	dummyURL := "http://example.com"

	invalidJSON := []byte(`{ "invalid": "json"`)
	req, _ := http.NewRequest(http.MethodPost, dummyURL, bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	_, err := getFunctionNameFromRequest(req)

	if err == nil {
		t.Errorf("Wanted an error but got nil")
		t.Fail()
	}
}

func createDeleteFunctionRequest(fn string, u string) *http.Request {
	delReq := requests.DeleteFunctionRequest{
		FunctionName: fn,
	}
	b, _ := json.Marshal(delReq)
	req, _ := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	return req

}

func Test_DeleteFunctionProxyHandlerSetsReplicasToZeroOnSuccessfulRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Success Response")
	}))
	defer ts.Close()

	metricsOptions := metrics.BuildMetricsOptions()
	u, _ := url.Parse(ts.URL)
	reverseProxy := types.NewHTTPClientReverseProxy(u, time.Second*5)
	notifiers := []HTTPNotifier{}
	urlResolver := SingleHostBaseURLResolver{BaseURL: ts.URL}
	h := MakeDeleteFunctionProxyHandler(reverseProxy, notifiers, urlResolver, metricsOptions)

	funcName := "test_function"
	req := createDeleteFunctionRequest(funcName, ts.URL)

	// Set metrics to 1 in order to confirm it was set to 0
	metricsOptions.
		ServiceReplicasCounter.
		WithLabelValues(funcName).Set(float64(1))

	w := httptest.NewRecorder()
	h(w, req)

	pb := &dto.Metric{}
	metricsOptions.
		ServiceReplicasCounter.
		WithLabelValues(funcName).
		Write(pb)

	replicaCount := pb.GetGauge().GetValue()
	if replicaCount != 0 {
		t.Logf("Resetting replicasCount failed, want: 0 got: %f", replicaCount)
		t.Fail()
	}

}

func Test_DeleteFunctionProxyHandlerDoesNotSetReplicasToZeroOnFailedRequest(t *testing.T) {

	metricsOptions := metrics.BuildMetricsOptions()
	fakeURL := "http://fake.url"
	u, _ := url.Parse(fakeURL)
	reverseProxy := types.NewHTTPClientReverseProxy(u, time.Second*5)
	notifiers := []HTTPNotifier{}
	urlResolver := SingleHostBaseURLResolver{BaseURL: fakeURL}
	h := MakeDeleteFunctionProxyHandler(reverseProxy, notifiers, urlResolver, metricsOptions)

	funcName := "test_function"
	req := createDeleteFunctionRequest(funcName, fakeURL)

	// Set metrics to 1 in order to confirm its value doesn't change on error
	metricsOptions.
		ServiceReplicasCounter.
		WithLabelValues(funcName).Set(float64(1))

	w := httptest.NewRecorder()
	h(w, req)

	pb := &dto.Metric{}
	metricsOptions.
		ServiceReplicasCounter.
		WithLabelValues(funcName).
		Write(pb)

	replicaCount := pb.GetGauge().GetValue()
	if replicaCount != 1.0 {
		t.Logf("ReplicasCount is wrong, want: 1.0 got: %f", replicaCount)
		t.Fail()
	}

}
