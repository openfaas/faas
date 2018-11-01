package plugin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas/gateway/scaling"
)

const fallbackValue = 120

func TestLabelValueWasEmpty(t *testing.T) {
	extractedValue := extractLabelValue("", fallbackValue)

	if extractedValue != fallbackValue {
		t.Log("Expected extractedValue to equal the fallbackValue")
		t.Fail()
	}
}

func TestLabelValueWasValid(t *testing.T) {
	extractedValue := extractLabelValue("42", fallbackValue)

	if extractedValue != 42 {
		t.Log("Expected extractedValue to equal answer to life (42)")
		t.Fail()
	}
}

func TestLabelValueWasInValid(t *testing.T) {
	extractedValue := extractLabelValue("InvalidValue", fallbackValue)

	if extractedValue != fallbackValue {
		t.Log("Expected extractedValue to equal the fallbackValue")
		t.Fail()
	}
}
func TestGetReplicasNonExistentFn(t *testing.T) {

	testServer := httptest.NewServer(
		http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusNotFound)
		}))
	defer testServer.Close()

	var creds auth.BasicAuthCredentials

	url, _ := url.Parse(testServer.URL + "/")

	esq := NewExternalServiceQuery(*url, &creds)

	svcQryResp, err := esq.GetReplicas("burt")

	if err == nil {
		t.Logf("Error was nil, expected non-nil - the service query response value was %+v ", svcQryResp)
		t.Fail()
	}
}

func TestGetReplicasExistentFn(t *testing.T) {

	testServer := httptest.NewServer(
		http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(`{"json":"body"}`))
		}))
	defer testServer.Close()

	expectedSvcQryResp := scaling.ServiceQueryResponse{
		Replicas:          0,
		MaxReplicas:       uint64(scaling.DefaultMaxReplicas),
		MinReplicas:       uint64(scaling.DefaultMinReplicas),
		ScalingFactor:     uint64(scaling.DefaultScalingFactor),
		AvailableReplicas: 0,
	}

	var creds auth.BasicAuthCredentials

	url, _ := url.Parse(testServer.URL + "/")

	esq := NewExternalServiceQuery(*url, &creds)

	svcQryResp, err := esq.GetReplicas("burt")

	if err != nil {
		t.Logf("Expected err to be nil got: %s ", err.Error())
		t.Fail()
	}
	if svcQryResp != expectedSvcQryResp {
		t.Logf("Unexpected return values - wanted %+v, got: %+v ", expectedSvcQryResp, svcQryResp)
		t.Fail()
	}
}

func TestSetReplicasNonExistentFn(t *testing.T) {

	testServer := httptest.NewServer(
		http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusInternalServerError)
		}))
	defer testServer.Close()

	var creds auth.BasicAuthCredentials
	url, _ := url.Parse(testServer.URL + "/")
	esq := NewExternalServiceQuery(*url, &creds)

	err := esq.SetReplicas("burt", 1)

	expectedErrStr := "error scaling HTTP code 500"

	if !strings.Contains(err.Error(), expectedErrStr) {
		t.Logf("Wanted string containing %s, got %s", expectedErrStr, err.Error())
		t.Fail()
	}
}

func TestSetReplicasExistentFn(t *testing.T) {

	testServer := httptest.NewServer(
		http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusOK)
		}))
	defer testServer.Close()

	var creds auth.BasicAuthCredentials
	url, _ := url.Parse(testServer.URL + "/")
	esq := NewExternalServiceQuery(*url, &creds)

	err := esq.SetReplicas("burt", 1)

	if err != nil {
		t.Logf("Expected err to be nil got: %s ", err.Error())
		t.Fail()
	}
}
