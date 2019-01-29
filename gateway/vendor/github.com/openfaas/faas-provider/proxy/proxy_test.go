package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
)

func varHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("name: %s params: %s", vars["name"], vars["params"])))
}

func testResolver(functionName string) (url.URL, error) {
	return url.URL{
		Scheme: "http",
		Host:   functionName,
	}, nil
}

func Test_pathParsing(t *testing.T) {
	tt := []struct {
		name         string
		functionPath string
		functionName string
		extraPath    string
		statusCode   int
	}{
		{
			"simple_name_match",
			"/function/echo",
			"echo",
			"",
			200,
		},
		{
			"simple_name_match_with_trailing_slash",
			"/function/echo/",
			"echo",
			"",
			200,
		},
		{
			"name_match_with_additional_path_values",
			"/function/echo/subPath/extras",
			"echo",
			"subPath/extras",
			200,
		},
		{
			"name_match_with_additional_path_values_and_querystring",
			"/function/echo/subPath/extras?query=true",
			"echo",
			"subPath/extras",
			200,
		},
		{
			"not_found_if_no_name",
			"/function/",
			"",
			"",
			404,
		},
	}

	// Need to create a router that we can pass the request through so that the vars will be added to the context
	router := mux.NewRouter()
	router.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", varHandler)
	router.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", varHandler)
	router.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/{params:.*}", varHandler)

	for _, s := range tt {
		t.Run(s.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", s.functionPath, nil)
			if err != nil {
				t.Fatal(err)
			}

			router.ServeHTTP(rr, req)
			if rr.Code != s.statusCode {
				t.Fatalf("unexpected status code; got: %d, expected: %d", rr.Code, s.statusCode)
			}

			body := rr.Body.String()
			expectedBody := fmt.Sprintf("name: %s params: %s", s.functionName, s.extraPath)
			if s.statusCode == http.StatusOK && body != expectedBody {
				t.Fatalf("incorrect function name and path params; got: %s, expected: %s", body, expectedBody)
			}
		})
	}
}

func Test_buildProxyRequest_Body_Method_Query(t *testing.T) {
	srcBytes := []byte("hello world")

	reader := bytes.NewReader(srcBytes)
	request, _ := http.NewRequest(http.MethodPost, "/?code=1", reader)
	request.Header.Set("X-Source", "unit-test")

	if request.URL.RawQuery != "code=1" {
		t.Errorf("Query - want: %s, got: %s", "code=1", request.URL.RawQuery)
		t.Fail()
	}

	funcURL, _ := testResolver("funcName")
	upstream, err := buildProxyRequest(request, funcURL, "")
	if err != nil {
		t.Fatal(err.Error())
	}

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

func Test_buildProxyRequest_NoBody_GetMethod_NoQuery(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/", nil)

	funcURL, _ := testResolver("funcName")
	upstream, err := buildProxyRequest(request, funcURL, "")
	if err != nil {
		t.Fatal(err.Error())
	}

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

func Test_buildProxyRequest_HasXForwardedHostHeaderWhenSet(t *testing.T) {
	srcBytes := []byte("hello world")

	reader := bytes.NewReader(srcBytes)
	request, err := http.NewRequest(http.MethodPost, "http://gateway/function?code=1", reader)

	if err != nil {
		t.Fatal(err)
	}

	funcURL, _ := testResolver("funcName")
	upstream, err := buildProxyRequest(request, funcURL, "/")
	if err != nil {
		t.Fatal(err.Error())
	}

	if request.Host != upstream.Header.Get("X-Forwarded-Host") {
		t.Errorf("Host - want: %s, got: %s", request.Host, upstream.Header.Get("X-Forwarded-Host"))
	}
}

func Test_buildProxyRequest_XForwardedHostHeader_Empty_WhenNotSet(t *testing.T) {
	srcBytes := []byte("hello world")

	reader := bytes.NewReader(srcBytes)
	request, err := http.NewRequest(http.MethodPost, "/function", reader)

	if err != nil {
		t.Fatal(err)
	}

	funcURL, _ := testResolver("funcName")
	upstream, err := buildProxyRequest(request, funcURL, "/")
	if err != nil {
		t.Fatal(err.Error())
	}

	if request.Host != upstream.Header.Get("X-Forwarded-Host") {
		t.Errorf("Host - want: %s, got: %s", request.Host, upstream.Header.Get("X-Forwarded-Host"))
	}
}

func Test_buildProxyRequest_XForwardedHostHeader_WhenAlreadyPresent(t *testing.T) {
	srcBytes := []byte("hello world")
	headerValue := "test.openfaas.com"
	reader := bytes.NewReader(srcBytes)
	request, err := http.NewRequest(http.MethodPost, "/function/test", reader)

	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("X-Forwarded-Host", headerValue)
	funcURL, _ := testResolver("funcName")
	upstream, err := buildProxyRequest(request, funcURL, "/")
	if err != nil {
		t.Fatal(err.Error())
	}

	if upstream.Header.Get("X-Forwarded-Host") != headerValue {
		t.Errorf("X-Forwarded-Host - want: %s, got: %s", headerValue, upstream.Header.Get("X-Forwarded-Host"))
	}
}

func Test_buildProxyRequest_WithPathNoQuery(t *testing.T) {
	srcBytes := []byte("hello world")
	functionPath := "/employee/info/300"

	requestPath := fmt.Sprintf("/function/xyz%s", functionPath)

	reader := bytes.NewReader(srcBytes)
	request, _ := http.NewRequest(http.MethodPost, requestPath, reader)
	request.Header.Set("X-Source", "unit-test")

	queryWant := ""
	if request.URL.RawQuery != queryWant {

		t.Errorf("Query - want: %s, got: %s", queryWant, request.URL.RawQuery)
		t.Fail()
	}

	funcURL, _ := testResolver("xyz")
	upstream, err := buildProxyRequest(request, funcURL, functionPath)
	if err != nil {
		t.Fatal(err.Error())
	}

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

	if functionPath != upstream.URL.Path {
		t.Errorf("URL.Path - want: %s, got: %s", functionPath, upstream.URL.Path)
		t.Fail()
	}

}

func Test_buildProxyRequest_WithNoPathNoQuery(t *testing.T) {
	srcBytes := []byte("hello world")
	functionPath := "/"

	requestPath := fmt.Sprintf("/function/xyz%s", functionPath)

	reader := bytes.NewReader(srcBytes)
	request, _ := http.NewRequest(http.MethodPost, requestPath, reader)
	request.Header.Set("X-Source", "unit-test")

	queryWant := ""
	if request.URL.RawQuery != queryWant {

		t.Errorf("Query - want: %s, got: %s", queryWant, request.URL.RawQuery)
		t.Fail()
	}

	funcURL, _ := testResolver("xyz")
	upstream, err := buildProxyRequest(request, funcURL, functionPath)
	if err != nil {
		t.Fatal(err.Error())
	}

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

	if functionPath != upstream.URL.Path {
		t.Errorf("URL.Path - want: %s, got: %s", functionPath, upstream.URL.Path)
		t.Fail()
	}

}

func Test_buildProxyRequest_WithPathAndQuery(t *testing.T) {
	srcBytes := []byte("hello world")
	functionPath := "/employee/info/300"

	requestPath := fmt.Sprintf("/function/xyz%s?code=1", functionPath)

	reader := bytes.NewReader(srcBytes)
	request, _ := http.NewRequest(http.MethodPost, requestPath, reader)
	request.Header.Set("X-Source", "unit-test")

	if request.URL.RawQuery != "code=1" {
		t.Errorf("Query - want: %s, got: %s", "code=1", request.URL.RawQuery)
		t.Fail()
	}

	funcURL, _ := testResolver("xyz")
	upstream, err := buildProxyRequest(request, funcURL, functionPath)
	if err != nil {
		t.Fatal(err.Error())
	}

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

	if functionPath != upstream.URL.Path {
		t.Errorf("URL.Path - want: %s, got: %s", functionPath, upstream.URL.Path)
		t.Fail()
	}

}
