// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"net/http"
	"testing"
)

func Test_Transform_RemovesFunctionPrefixRootPath(t *testing.T) {

	req, _ := http.NewRequest(http.MethodGet, "/function/figlet", nil)
	transformer := FunctionPrefixTrimmingURLPathTransformer{}
	want := ""
	got := transformer.Transform(req)

	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func Test_Transform_RemovesFunctionPrefixWithSingleParam(t *testing.T) {

	req, _ := http.NewRequest(http.MethodGet, "/function/figlet/employees", nil)
	transformer := FunctionPrefixTrimmingURLPathTransformer{}
	want := "/employees"
	got := transformer.Transform(req)

	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func Test_Transform_RemovesFunctionPrefixWithDotInName(t *testing.T) {

	req, _ := http.NewRequest(http.MethodGet, "/function/figlet.fn", nil)
	transformer := FunctionPrefixTrimmingURLPathTransformer{}
	want := ""
	got := transformer.Transform(req)

	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func Test_Transform_RemovesFunctionPrefixWithDotInNameAndPath(t *testing.T) {

	req, _ := http.NewRequest(http.MethodGet, "/function/figlet.fn/employees", nil)
	transformer := FunctionPrefixTrimmingURLPathTransformer{}
	want := "/employees"
	got := transformer.Transform(req)

	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func Test_Transform_RemovesFunctionPrefixWithParams(t *testing.T) {

	req, _ := http.NewRequest(http.MethodGet, "/function/figlet/employees/100", nil)
	transformer := FunctionPrefixTrimmingURLPathTransformer{}
	want := "/employees/100"
	got := transformer.Transform(req)

	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}
}
