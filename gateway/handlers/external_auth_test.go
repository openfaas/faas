package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_External_Auth_Wrapper_FailsInvalidAuth(t *testing.T) {

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer s.Close()

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}

	passBody := false
	handler := MakeExternalAuthHandler(next, time.Second*5, s.URL, passBody)

	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code == http.StatusOK {
		t.Errorf("Status incorrect, did not want: %d, but got %d", http.StatusOK, rr.Code)
	}
}

func Test_External_Auth_Wrapper_PassesValidAuth(t *testing.T) {

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}

	passBody := false
	handler := MakeExternalAuthHandler(next, time.Second*5, s.URL, passBody)

	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	rr := httptest.NewRecorder()
	handler(rr, req)
	want := http.StatusNotImplemented
	if rr.Code != want {
		t.Errorf("Status incorrect, want: %d, but got %d", want, rr.Code)
	}
}

func Test_External_Auth_Wrapper_TimeoutGivesInternalServerError(t *testing.T) {

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}

	passBody := false
	handler := MakeExternalAuthHandler(next, time.Millisecond*10, s.URL, passBody)

	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	want := http.StatusInternalServerError
	if rr.Code != want {
		t.Errorf("Status incorrect, want: %d, but got %d", want, rr.Code)
	}
}

// // Test_External_Auth_Wrapper_PassesValidAuthButOnly200IsValid this test exists
// // to document the TODO action to consider all "2xx" statuses as valid.
// func Test_External_Auth_Wrapper_PassesValidAuthButOnly200IsValid(t *testing.T) {

// 	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusAccepted)
// 	}))
// 	defer s.Close()

// 	next := func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusNotImplemented)
// 	}

// 	passBody := false
// 	handler := MakeExternalAuthHandler(next, time.Second*5, s.URL, passBody)

// 	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
// 	rr := httptest.NewRecorder()
// 	handler(rr, req)
// 	want := http.StatusUnauthorized
// 	if rr.Code != want {
// 		t.Errorf("Status incorrect, want: %d, but got %d", want, rr.Code)
// 	}
// }
