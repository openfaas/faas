// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"net/http"
	"testing"
)

func Test_getNameParts(t *testing.T) {
	fn, ns := getNameParts("figlet.openfaas-fn")
	wantFn := "figlet"
	wantNs := "openfaas-fn"

	if fn != wantFn {
		t.Fatalf("want %s, got %s", wantFn, fn)
	}
	if ns != wantNs {
		t.Fatalf("want %s, got %s", wantNs, ns)
	}
}

func Test_getNamePartsDualDot(t *testing.T) {
	fn, ns := getNameParts("dev.figlet.openfaas-fn")
	wantFn := "dev.figlet"
	wantNs := "openfaas-fn"

	if fn != wantFn {
		t.Fatalf("want %s, got %s", wantFn, fn)
	}
	if ns != wantNs {
		t.Fatalf("want %s, got %s", wantNs, ns)
	}
}

func Test_getNameParts_NoNs(t *testing.T) {
	fn, ns := getNameParts("figlet")
	wantFn := "figlet"
	wantNs := ""

	if fn != wantFn {
		t.Fatalf("want %s, got %s", wantFn, fn)
	}
	if ns != wantNs {
		t.Fatalf("want %s, got %s", wantNs, ns)
	}
}

func Test_getCallbackURLHeader(t *testing.T) {
	want := "http://localhost:8080"
	header := http.Header{}
	header.Add("X-Callback-Url", want)

	uri, err := getCallbackURLHeader(header)
	if err != nil {
		t.Fatal(err)
	}

	if uri.String() != want {
		t.Fatalf("want %s, but got %s", want, uri.String())
	}
}

func Test_getCallbackURLHeader_ParseFails(t *testing.T) {
	want := "ht tp://foo.com"
	header := http.Header{}
	header.Add("X-Callback-Url", want)

	_, err := getCallbackURLHeader(header)
	if err == nil {
		t.Fatal("wanted a parsing error.")
	}
}
