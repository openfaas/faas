// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHandler_make(t *testing.T) {
	config := WatchdogConfig{}
	handler := makeRequestHandler(&config)

	if handler == nil {
		t.Fail()
	}
}

func TestHandler_HasCustomHeaderInFunction_WithCgi_Mode(t *testing.T) {
	rr := httptest.NewRecorder()

	body := ""
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Add("custom-header", "value")

	if err != nil {
		t.Fatal(err)
	}

	config := WatchdogConfig{
		faasProcess: "env",
		cgiHeaders:  true,
	}
	handler := makeRequestHandler(&config)
	handler(rr, req)

	required := http.StatusOK
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code - got: %v, want: %v",
			status, required)
	}

	read, _ := ioutil.ReadAll(rr.Body)
	val := string(read)
	if !strings.Contains(val, "Http_ContentLength=0") {
		t.Errorf(config.faasProcess+" should print: Http_ContentLength=0, got: %s\n", val)
	}
	if !strings.Contains(val, "Http_Custom_Header") {
		t.Errorf(config.faasProcess+" should print: Http_Custom_Header, got: %s\n", val)
	}

	seconds := rr.Header().Get("X-Duration-Seconds")
	if len(seconds) == 0 {
		t.Errorf(config.faasProcess + " should have given a duration as an X-Duration-Seconds header\n")
	}
}

func TestHandler_HasCustomHeaderInFunction_WithCgiMode_AndBody(t *testing.T) {
	rr := httptest.NewRecorder()

	body := "test"
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Add("custom-header", "value")

	if err != nil {
		t.Fatal(err)
	}

	config := WatchdogConfig{
		faasProcess: "env",
		cgiHeaders:  true,
	}
	handler := makeRequestHandler(&config)
	handler(rr, req)

	required := http.StatusOK
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code - got: %v, want: %v",
			status, required)
	}

	read, _ := ioutil.ReadAll(rr.Body)
	val := string(read)
	if !strings.Contains(val, fmt.Sprintf("Http_ContentLength=%d", len(body))) {
		t.Errorf("'env' should printed: Http_ContentLength=0, got: %s\n", val)
	}
	if !strings.Contains(val, "Http_Custom_Header") {
		t.Errorf("'env' should printed: Http_Custom_Header, got: %s\n", val)
	}

	seconds := rr.Header().Get("X-Duration-Seconds")
	if len(seconds) == 0 {
		t.Errorf("Exec of cat should have given a duration as an X-Duration-Seconds header\n")
	}
}

func TestHandler_StderrWritesToStderr_CombinedOutput_False(t *testing.T) {
	rr := httptest.NewRecorder()

	b := bytes.NewBuffer([]byte{})
	log.SetOutput(b)

	body := ""
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

	if err != nil {
		t.Fatal(err)
	}

	config := WatchdogConfig{
		faasProcess:   "stat x",
		cgiHeaders:    true,
		combineOutput: false,
	}

	handler := makeRequestHandler(&config)
	handler(rr, req)

	required := http.StatusInternalServerError

	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code - got: %v, want: %v",
			status, required)
	}

	log.SetOutput(os.Stderr)

	stderrBytes, _ := ioutil.ReadAll(b)
	stderrVal := string(stderrBytes)

	want := "No such file or directory"
	if strings.Contains(stderrVal, want) == false {
		t.Logf("Stderr should have contained error from function \"%s\", but was: %s", want, stderrVal)
		t.Fail()
	}
}

func TestHandler_StderrWritesToResponse_CombinedOutput_True(t *testing.T) {
	rr := httptest.NewRecorder()

	b := bytes.NewBuffer([]byte{})
	log.SetOutput(b)

	body := ""
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

	if err != nil {
		t.Fatal(err)
	}

	config := WatchdogConfig{
		faasProcess:   "stat x",
		cgiHeaders:    true,
		combineOutput: true,
	}

	handler := makeRequestHandler(&config)
	handler(rr, req)

	required := http.StatusInternalServerError

	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code - got: %v, want: %v",
			status, required)
	}

	log.SetOutput(os.Stderr)

	stderrBytes, _ := ioutil.ReadAll(b)
	stderrVal := string(stderrBytes)
	stdErrWant := "No such file or directory"
	if strings.Contains(stderrVal, stdErrWant) {
		t.Logf("stderr should have not included any function errors, but did")
		t.Fail()
	}

	bodyBytes, _ := ioutil.ReadAll(rr.Body)
	bodyStr := string(bodyBytes)
	stdOuputWant := `exit status 1`
	if strings.Contains(bodyStr, stdOuputWant) == false {
		t.Logf("response want: %s, got: %s", stdOuputWant, bodyStr)
		t.Fail()
	}
	if strings.Contains(bodyStr, stdErrWant) == false {
		t.Logf("response want: %s, got: %s", stdErrWant, bodyStr)
		t.Fail()
	}
}

func TestHandler_DoesntHaveCustomHeaderInFunction_WithoutCgi_Mode(t *testing.T) {
	rr := httptest.NewRecorder()

	body := ""
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Add("custom-header", "value")
	if err != nil {
		t.Fatal(err)
	}

	config := WatchdogConfig{
		faasProcess: "env",
		cgiHeaders:  false,
	}
	handler := makeRequestHandler(&config)
	handler(rr, req)

	required := http.StatusOK
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code - got: %v, want: %v",
			status, required)
	}

	read, _ := ioutil.ReadAll(rr.Body)
	val := string(read)
	if strings.Contains(val, "Http_Custom_Header") {
		t.Errorf("'env' should not have printed: Http_Custom_Header, got: %s\n", val)

	}

	seconds := rr.Header().Get("X-Duration-Seconds")
	if len(seconds) == 0 {
		t.Errorf("Exec of cat should have given a duration as an X-Duration-Seconds header\n")
	}
}

func TestHandler_HasXDurationSecondsHeader(t *testing.T) {
	rr := httptest.NewRecorder()

	body := "hello"
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
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
		t.Errorf("handler returned wrong status code - got: %v, want: %v",
			status, required)
	}

	seconds := rr.Header().Get("X-Duration-Seconds")
	if len(seconds) == 0 {
		t.Errorf("Exec of " + config.faasProcess + " should have given a duration as an X-Duration-Seconds header")
	}
}

func TestHandler_RequestTimeoutFailsForExceededDuration(t *testing.T) {
	rr := httptest.NewRecorder()

	verbs := []string{http.MethodPost}
	for _, verb := range verbs {

		body := "hello"
		req, err := http.NewRequest(verb, "/", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}

		config := WatchdogConfig{
			faasProcess: "sleep 2",
			execTimeout: time.Duration(100) * time.Millisecond,
		}

		handler := makeRequestHandler(&config)
		handler(rr, req)

		required := http.StatusRequestTimeout
		if status := rr.Code; status != required {
			t.Errorf("handler returned wrong status code for verb [%s] - got: %v, want: %v",
				verb, status, required)
		}
	}
}

func TestHandler_StatusOKAllowed_ForWriteableVerbs(t *testing.T) {
	rr := httptest.NewRecorder()

	verbs := []string{http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, verb := range verbs {

		body := "hello"
		req, err := http.NewRequest(verb, "/", bytes.NewBufferString(body))
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
			t.Errorf("handler returned wrong status code for verb [%s] - got: %v, want: %v",
				verb, status, required)
		}

		buf, _ := ioutil.ReadAll(rr.Body)
		val := string(buf)
		if val != body {
			t.Errorf("Exec of cat did not return input value, %s", val)
		}
	}
}

func TestHandler_StatusMethodNotAllowed_ForUnknown(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("UNKNOWN", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	config := WatchdogConfig{}
	handler := makeRequestHandler(&config)
	handler(rr, req)

	required := http.StatusMethodNotAllowed
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code: got %v, want: %v",
			status, required)
	}
}

func TestHandler_StatusOKForGETAndNoBody(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	config := WatchdogConfig{
		// writeDebug:  true,
		faasProcess: "date",
	}

	handler := makeRequestHandler(&config)
	handler(rr, req)

	required := http.StatusOK
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code: got %v, want: %v",
			status, required)
	}
}

func TestHealthHandler_StatusOK_LockFilePresent(t *testing.T) {
	rr := httptest.NewRecorder()

	present := lockFilePresent()

	if present == false {
		if _, err := createLockFile(); err != nil {
			t.Fatal(err)
		}
	}

	req, err := http.NewRequest(http.MethodGet, "/_/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler := makeHealthHandler()
	handler(rr, req)

	required := http.StatusOK
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code: got %v, but wanted %v", status, required)
	}

}

func TestHealthHandler_StatusInternalServerError_LockFileNotPresent(t *testing.T) {
	rr := httptest.NewRecorder()

	if lockFilePresent() == true {
		if err := removeLockFile(); err != nil {
			t.Fatal(err)
		}
	}

	req, err := http.NewRequest(http.MethodGet, "/_/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler := makeHealthHandler()
	handler(rr, req)

	required := http.StatusInternalServerError
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code - got: %v, want: %v", status, required)
	}
}

func TestHealthHandler_SatusMethoNotAllowed_ForWriteableVerbs(t *testing.T) {
	rr := httptest.NewRecorder()

	verbs := []string{http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, verb := range verbs {
		req, err := http.NewRequest(verb, "/_/health", nil)
		if err != nil {
			t.Fatal(err)
		}

		handler := makeHealthHandler()
		handler(rr, req)

		required := http.StatusMethodNotAllowed
		if status := rr.Code; status != required {
			t.Errorf("handler returned wrong status code: got %v, but wanted %v", status, required)
		}
	}
}

func removeLockFile() error {
	path := filepath.Join(os.TempDir(), ".lock")
	log.Printf("Removing lock-file : %s\n", path)
	removeErr := os.Remove(path)
	return removeErr
}
