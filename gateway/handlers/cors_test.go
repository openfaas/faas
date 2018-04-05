package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type customHandler struct {
}

func (h customHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func Test_HeadersAdded(t *testing.T) {
	rr := httptest.NewRecorder()
	handler := customHandler{}
	host := "store.openfaas.com"

	decorated := DecorateWithCORS(handler, host)
	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	decorated.ServeHTTP(rr, request)

	actual := rr.Header().Get("Access-Control-Allow-Origin")
	if actual != host {
		t.Errorf("Access-Control-Allow-Origin: want: %s got: %s", host, actual)
	}

	actualMethods := rr.Header().Get("Access-Control-Allow-Methods")
	if actualMethods != http.MethodGet {
		t.Errorf("Access-Control-Allow-Methods: want: %s got: %s", http.MethodGet, actualMethods)
	}

}
