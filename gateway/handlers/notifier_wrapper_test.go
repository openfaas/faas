// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.

package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_MakeNotifierWrapper_ReceivesHttpStatusInNotifier(t *testing.T) {
	notifier := &testNotifier{}
	handlerVisited := false
	handlerWant := http.StatusAccepted

	handler := MakeNotifierWrapper(func(w http.ResponseWriter, r *http.Request) {
		handlerVisited = true
		w.WriteHeader(handlerWant)
	}, []HTTPNotifier{notifier})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if handlerVisited != true {
		t.Errorf("expected handler to have been visited")
	}

	if notifier.StatusReceived == 0 {
		t.Errorf("notifier wanted a status, but got none")
		t.Fail()
		return
	}

	if rec.Result().StatusCode != handlerWant {
		t.Errorf("recorder status want: %d, got %d", handlerWant, rec.Result().StatusCode)
	}

	if notifier.StatusReceived != handlerWant {
		t.Errorf("notifier status want: %d, got %d", handlerWant, notifier.StatusReceived)
	}

}

func Test_MakeNotifierWrapper_ReceivesDefaultHttpStatusWhenNotSet(t *testing.T) {
	notifier := &testNotifier{}
	handlerVisited := false
	handlerWant := http.StatusOK

	handler := MakeNotifierWrapper(func(w http.ResponseWriter, r *http.Request) {
		handlerVisited = true
	}, []HTTPNotifier{notifier})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if handlerVisited != true {
		t.Errorf("expected handler to have been visited")
	}

	if notifier.StatusReceived == 0 {
		t.Errorf("notifier wanted a status, but got none")
		t.Fail()
		return
	}

	if rec.Result().StatusCode != handlerWant {
		t.Errorf("recorder status want: %d, got %d", handlerWant, rec.Result().StatusCode)
	}

	if notifier.StatusReceived != handlerWant {
		t.Errorf("notifier status want: %d, got %d", handlerWant, notifier.StatusReceived)
	}

}

type testNotifier struct {
	StatusReceived int
}

// Notify about service metrics
func (tf *testNotifier) Notify(method string, URL string, originalURL string, statusCode int, event string, duration time.Duration) {
	tf.StatusReceived = statusCode
}

func TestLoggingMiddleware(t *testing.T) {

	logger := LoggingNotifier{}
	cases := []struct {
		name   string
		status int
		method string
		path   string
	}{
		{
			name:   "logs successful GET request",
			status: http.StatusOK,
			method: http.MethodGet,
			path:   "/a/b/c",
		},
		{
			name:   "logs successful POST request",
			status: http.StatusOK,
			method: http.MethodPost,
			path:   "/a/b/c",
		},
		{
			name:   "logs successful PATCH request",
			status: http.StatusOK,
			method: http.MethodPatch,
			path:   "/a/b/c",
		},
		{
			name:   "logs successful PUT request",
			status: http.StatusOK,
			method: http.MethodPut,
			path:   "/a/b/c",
		},
		{
			name:   "logs successful DELETE request",
			status: http.StatusOK,
			method: http.MethodDelete,
			path:   "/a/b/c",
		},
		{
			name:   "logs successful OPTIONS request",
			status: http.StatusOK,
			method: http.MethodOptions,
			path:   "/a/b/c",
		},
		{
			name:   "logs 201 success",
			status: http.StatusCreated,
			method: http.MethodGet,
			path:   "/a/b/c",
		},
		{
			name:   "logs 204 success",
			status: http.StatusNoContent,
			method: http.MethodGet,
			path:   "/a/b/c",
		},
		{
			name:   "logs 400 failure",
			status: http.StatusBadRequest,
			method: http.MethodGet,
			path:   "/a/b/c",
		},
		{
			name:   "logs 401 failure",
			status: http.StatusUnauthorized,
			method: http.MethodGet,
			path:   "/a/b/c",
		},
		{
			name:   "logs 403 failure",
			status: http.StatusForbidden,
			method: http.MethodGet,
			path:   "/a/b/c",
		},
		{
			name:   "logs 500 failure",
			status: http.StatusInternalServerError,
			method: http.MethodGet,
			path:   "/a/b/c",
		},
		{
			name:   "logs 301 redirect",
			status: http.StatusMovedPermanently,
			method: http.MethodGet,
			path:   "/a/b/c",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			log.SetOutput(&b)
			log.SetFlags(0)
			log.SetPrefix("")

			handler := MakeNotifierWrapper(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			}, []HTTPNotifier{logger})

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("unexpected status code, expected %d, got %d", tc.status, rec.Code)
			}

			logs := b.String()

			prefix := fmt.Sprintf("Forwarded [%s] to %s - [%d] -", tc.method, tc.path, tc.status)
			if !strings.HasPrefix(logs, prefix) {
				t.Fatalf("expected log to start with: %q\ngot: %q", prefix, logs)
			}
		})
	}

}
