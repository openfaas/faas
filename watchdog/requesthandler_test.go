// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_make(t *testing.T) {
	config := WatchdogConfig{}
	handler := makeRequestHandler(&config)

	if handler == nil {
		t.Fail()
	}
}

func TestHandler_HasXDurationSecondsHeader(t *testing.T) {
	rr := httptest.NewRecorder()

	body := "hello"
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}

	config := WatchdogConfig{
		faasProcess: "cat",
	}
	handler := makeRequestHandler(&config)
	handler(rr, req)

	required := http.StatusOK
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code: got %v, but wanted %v",
			status, required)
	}

	seconds := rr.Header().Get("X-Duration-Seconds")
	if len(seconds) == 0 {
		t.Errorf("Exec of cat should have given a duration as an X-Duration-Seconds header\n")
	}
}

func TestHandler_StatusOKAllowed_ForPOST(t *testing.T) {
	rr := httptest.NewRecorder()

	body := "hello"
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}

	config := WatchdogConfig{
		faasProcess: "cat",
	}
	handler := makeRequestHandler(&config)
	handler(rr, req)

	required := http.StatusOK
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code: got %v, but wanted %v",
			status, required)
	}

	buf, _ := ioutil.ReadAll(rr.Body)
	val := string(buf)
	if val != body {
		t.Errorf("Exec of cat did not return input value, %s", val)
	}
}

func TestHandler_StatusMethodNotAllowed_ForGet(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	config := WatchdogConfig{}
	handler := makeRequestHandler(&config)
	handler(rr, req)

	required := http.StatusMethodNotAllowed
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code: got %v, but wanted %v",
			status, required)
	}
}
