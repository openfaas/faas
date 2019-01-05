// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"net/http"
	"net/http/httptest"
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
func (tf *testNotifier) Notify(method string, URL string, originalURL string, statusCode int, duration time.Duration) {
	tf.StatusReceived = statusCode
}
