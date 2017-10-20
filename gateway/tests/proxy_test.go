package tests

import (
	"net/http"
	"testing"

	"github.com/openfaas/faas/gateway/handlers"
)

func Test_GetContentType_UsesResponseValue(t *testing.T) {
	request := http.Header{}
	request.Add("Content-Type", "text/plain")
	response := http.Header{}
	response.Add("Content-Type", "text/html")

	contentType := handlers.GetContentType(request, response, "default")
	if contentType != response.Get("Content-Type") {
		t.Errorf("Got: %s, want: %s", contentType, response.Get("Content-Type"))
	}
}

func Test_GetContentType_UsesRequest_WhenResponseEmpty(t *testing.T) {
	request := http.Header{}
	request.Add("Content-Type", "text/plain")
	response := http.Header{}
	response.Add("Content-Type", "")

	contentType := handlers.GetContentType(request, response, "default")
	if contentType != request.Get("Content-Type") {
		t.Errorf("Got: %s, want: %s", contentType, request.Get("Content-Type"))
	}

}

func Test_GetContentType_UsesDefaultWhenRequestResponseEmpty(t *testing.T) {
	request := http.Header{}
	request.Add("Content-Type", "")
	response := http.Header{}
	response.Add("Content-Type", "")

	contentType := handlers.GetContentType(request, response, "default")
	if contentType != "default" {
		t.Errorf("Got: %s, want: %s", contentType, "default")
	}

}
