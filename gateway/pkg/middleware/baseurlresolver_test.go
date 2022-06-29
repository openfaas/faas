// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package middleware

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func Test_SingleHostBaseURLResolver_BuildURL(t *testing.T) {

	newNamespace := "production-fn"
	function := "figlet"
	r := SingleHostBaseURLResolver{BaseURL: "http://faas-netes.openfaas:8080"}

	want := "http://faas-netes.openfaas:8080/function/figlet.production-fn/healthz"

	got := r.BuildURL(function, newNamespace, "/healthz", true)
	if got != want {
		t.Fatalf("r.URL failed, want: %s got: %s", want, got)
	}
}

func Test_SingleHostBaseURLResolver_BuildURL_DefaultNamespace(t *testing.T) {

	newNamespace := "openfaas-fn"
	function := "figlet"
	r := SingleHostBaseURLResolver{BaseURL: "http://faas-netes.openfaas:8080"}

	want := "http://faas-netes.openfaas:8080/function/figlet.openfaas-fn/_/health"

	got := r.BuildURL(function, newNamespace, "/_/health", true)
	if got != want {
		t.Fatalf("r.URL failed, want: %s got: %s", want, got)
	}
}

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

func TestURL_NonDefaultNamespaceWithPath(t *testing.T) {
	suffix := "openfaas-fn.local.cluster.svc"
	namespace := "openfaas-fn"
	newNamespace := "production-fn"
	function := "figlet"
	r := FunctionAsHostBaseURLResolver{FunctionSuffix: suffix, FunctionNamespace: namespace}

	want := "http://figlet.production-fn.local.cluster.svc:8080/healthz"

	got := r.BuildURL(function, newNamespace, "/healthz", true)
	if got != want {
		t.Fatalf("r.URL failed, want: %s got: %s", want, got)
	}
}

func TestURL_NonDefaultNamespaceWithout(t *testing.T) {
	suffix := "openfaas-fn.local.cluster.svc"
	namespace := "openfaas-fn"
	newNamespace := "production-fn"
	function := "figlet"
	r := FunctionAsHostBaseURLResolver{FunctionSuffix: suffix, FunctionNamespace: namespace}

	want := "http://figlet.production-fn.local.cluster.svc:8080"

	got := r.BuildURL(function, newNamespace, "", true)
	if got != want {
		t.Fatalf("r.URL failed, want: %s got: %s", want, got)
	}
}

func TestURL_DefaultNamespaceWithPath(t *testing.T) {
	suffix := "openfaas-fn.local.cluster.svc"
	namespace := "openfaas-fn"
	newNamespace := "production-fn"
	function := "figlet"
	r := FunctionAsHostBaseURLResolver{FunctionSuffix: suffix, FunctionNamespace: namespace}

	want := "http://figlet.production-fn.local.cluster.svc:8080/_/health"

	got := r.BuildURL(function, newNamespace, "/_/health", true)
	if got != want {
		t.Fatalf("r.URL failed, want: %s got: %s", want, got)
	}
}

func TestFunctionAsHostBaseURLResolver_WithNamespaceOverride(t *testing.T) {

	suffix := "openfaas-fn.local.cluster.svc."
	namespace := "openfaas-fn"
	newNS := "production-fn"

	r := FunctionAsHostBaseURLResolver{FunctionSuffix: suffix, FunctionNamespace: namespace}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/function/hello."+newNS, nil)

	resolved := r.Resolve(req)

	newSuffix := strings.Replace(suffix, namespace, newNS, -1)

	want := fmt.Sprintf("http://hello.%s:%d", newSuffix, watchdogPort)
	log.Println(want)
	if resolved != want {
		t.Logf("r.Resolve failed, want: %s got: %s", want, resolved)
		t.Fail()
	}
}

func TestFunctionAsHostBaseURLResolver_WithSuffix(t *testing.T) {
	suffix := "openfaas-fn.local.cluster.svc."
	r := FunctionAsHostBaseURLResolver{FunctionSuffix: suffix}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/function/hello", nil)

	resolved := r.Resolve(req)
	want := fmt.Sprintf("http://hello.%s:%d", suffix, watchdogPort)
	log.Println(want)
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
