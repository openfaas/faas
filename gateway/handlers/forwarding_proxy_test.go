package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func Test_buildUpstreamRequest_Body_Method_Query(t *testing.T) {
	srcBytes := []byte("hello world")

	reader := bytes.NewReader(srcBytes)
	request, _ := http.NewRequest(http.MethodPost, "/?code=1", reader)
	request.Header.Set("X-Source", "unit-test")

	if request.URL.RawQuery != "code=1" {
		t.Errorf("Query - want: %s, got: %s", "code=1", request.URL.RawQuery)
		t.Fail()
	}

	upstream := buildUpstreamRequest(request, "/")

	if request.Method != upstream.Method {
		t.Errorf("Method - want: %s, got: %s", request.Method, upstream.Method)
		t.Fail()
	}

	upstreamBytes, _ := ioutil.ReadAll(upstream.Body)

	if string(upstreamBytes) != string(srcBytes) {
		t.Errorf("Body - want: %s, got: %s", string(upstreamBytes), string(srcBytes))
		t.Fail()
	}

	if request.Header.Get("X-Source") != upstream.Header.Get("X-Source") {
		t.Errorf("Header X-Source - want: %s, got: %s", request.Header.Get("X-Source"), upstream.Header.Get("X-Source"))
		t.Fail()
	}

	if request.URL.RawQuery != upstream.URL.RawQuery {
		t.Errorf("URL.RawQuery - want: %s, got: %s", request.URL.RawQuery, upstream.URL.RawQuery)
		t.Fail()
	}

}

func Test_buildUpstreamRequest_NoBody_GetMethod_NoQuery(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/", nil)

	upstream := buildUpstreamRequest(request, "/")

	if request.Method != upstream.Method {
		t.Errorf("Method - want: %s, got: %s", request.Method, upstream.Method)
		t.Fail()
	}

	if upstream.Body != nil {
		t.Errorf("Body - expected nil")
		t.Fail()
	}

	if request.URL.RawQuery != upstream.URL.RawQuery {
		t.Errorf("URL.RawQuery - want: %s, got: %s", request.URL.RawQuery, upstream.URL.RawQuery)
		t.Fail()
	}

}

func Test_getServiceName(t *testing.T) {
	scenarios := []struct {
		name        string
		url         string
		serviceName string
	}{
		{
			name:        "can handle request without trailing slash",
			url:         "/function/testFunc",
			serviceName: "testFunc",
		},
		{
			name:        "can handle request with trailing slash",
			url:         "/function/testFunc/",
			serviceName: "testFunc",
		},
		{
			name:        "can handle request with query parameters",
			url:         "/function/testFunc?name=foo",
			serviceName: "testFunc",
		},
		{
			name:        "can handle request with trailing slash and query parameters",
			url:         "/function/testFunc/?name=foo",
			serviceName: "testFunc",
		},
		{
			name:        "can handle request with a fragment",
			url:         "/function/testFunc#fragment",
			serviceName: "testFunc",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {

			u, err := url.Parse("http://openfaas.local" + s.url)
			if err != nil {
				t.Fatal(err)
			}

			service := getServiceName(u.Path)
			if service != s.serviceName {
				t.Fatalf("Incorrect service name - want: %s, got: %s", s.serviceName, service)
			}
		})
	}
}
