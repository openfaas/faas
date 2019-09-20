// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
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
