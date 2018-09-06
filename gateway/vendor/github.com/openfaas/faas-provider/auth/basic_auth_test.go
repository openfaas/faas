// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_AuthWithValidPassword_Gives200(t *testing.T) {

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "<html><body>Hello World!</body></html>")
	}
	w := httptest.NewRecorder()

	wantUser := "admin"
	wantPassword := "password"
	r := httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	r.SetBasicAuth(wantUser, wantPassword)
	wantCredentials := &BasicAuthCredentials{
		User:     wantUser,
		Password: wantPassword,
	}

	decorated := DecorateWithBasicAuth(handler, wantCredentials)
	decorated.ServeHTTP(w, r)

	wantCode := http.StatusOK

	if w.Code != wantCode {
		t.Errorf("status code, want: %d, got: %d", wantCode, w.Code)
		t.Fail()
	}
}

func Test_AuthWithInvalidPassword_Gives403(t *testing.T) {

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "<html><body>Hello World!</body></html>")
	}

	w := httptest.NewRecorder()

	wantUser := "admin"
	wantPassword := "test"
	r := httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	r.SetBasicAuth(wantUser, wantPassword)

	wantCredentials := &BasicAuthCredentials{
		User:     wantUser,
		Password: "",
	}

	decorated := DecorateWithBasicAuth(handler, wantCredentials)
	decorated.ServeHTTP(w, r)

	wantCode := http.StatusUnauthorized
	if w.Code != wantCode {
		t.Errorf("status code, want: %d, got: %d", wantCode, w.Code)
		t.Fail()
	}
}
