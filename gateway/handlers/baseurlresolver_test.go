package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestSingleHostBaseURLResolver(t *testing.T) {

	urlVal, _ := url.Parse("http://upstream:8080/")
	r := SingleHostBaseURLResolver{BaseURL: urlVal.String()}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/function/hello", nil)

	resolved := r.Resolve(req)
	want := "http://upstream:8080"
	if resolved != want {
		t.Logf("r.Resolve failed, want: %s got: %s", want, resolved)
		t.Fail()
	}
}

const watchdogPort = 8080

func TestFunctionAsHostBaseURLResolver_WithSuffix(t *testing.T) {
	suffix := "openfaas-fn.local.cluster.svc."
	r := FunctionAsHostBaseURLResolver{FunctionSuffix: suffix}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/function/hello", nil)

	resolved := r.Resolve(req)
	want := fmt.Sprintf("http://hello.%s:%d", suffix, watchdogPort)

	if resolved != want {
		t.Logf("r.Resolve failed, want: %s got: %s", want, resolved)
		t.Fail()
	}
}

func TestFunctionAsHostBaseURLResolver_WithoutSuffix(t *testing.T) {
	suffix := ""
	r := FunctionAsHostBaseURLResolver{FunctionSuffix: suffix}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/function/hello", nil)

	resolved := r.Resolve(req)
	want := fmt.Sprintf("http://hello%s:%d", suffix, watchdogPort)

	if resolved != want {
		t.Logf("r.Resolve failed, want: %s got: %s", want, resolved)
		t.Fail()
	}
}
