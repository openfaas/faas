package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
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

func Test_External_Auth_Wrapper_FailsInvalidAuth_WritesBody(t *testing.T) {

	wantBody := []byte(`invalid credentials`)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write(wantBody)
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

	if bytes.Compare(rr.Body.Bytes(), wantBody) != 0 {
		t.Errorf("Body incorrect, want: %s, but got %s", []byte(wantBody), rr.Body)
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

func Test_External_Auth_Wrapper_WithoutRequiredHeaderFailsAuth(t *testing.T) {
	wantToken := "secret-key"
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") == wantToken {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer s.Close()

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}

	passBody := false
	handler := MakeExternalAuthHandler(next, time.Second*5, s.URL, passBody)

	req := httptest.NewRequest(http.MethodGet, s.URL, nil)

	// use an invalid token
	req.Header.Set("X-Token", "invalid-key")

	rr := httptest.NewRecorder()
	handler(rr, req)
	want := http.StatusUnauthorized
	if rr.Code != want {
		t.Errorf("Status incorrect, want: %d, but got %d", want, rr.Code)
	}
}

func Test_External_Auth_Wrapper_WithoutRequiredHeaderFailsAuth_ProxiesServerHeaders(t *testing.T) {
	wantToken := "secret-key"
	wantRealm := `Basic realm="Restricted"`
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") == wantToken {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Www-Authenticate", wantRealm)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer s.Close()

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}

	passBody := false
	handler := MakeExternalAuthHandler(next, time.Second*5, s.URL, passBody)

	req := httptest.NewRequest(http.MethodGet, s.URL, nil)

	// use an invalid token
	req.Header.Set("X-Token", "invalid-key")

	rr := httptest.NewRecorder()
	handler(rr, req)
	want := http.StatusUnauthorized
	if rr.Code != want {
		t.Errorf("Status incorrect, want: %d, but got %d", want, rr.Code)
	}

	got := rr.Header().Get("Www-Authenticate")
	if got != wantRealm {
		t.Errorf("Www-Authenticate header, want: %s, but got %s, %q", wantRealm, got, rr.Header())
	}
}

func Test_External_Auth_Wrapper_WithRequiredHeaderPassesValidAuth(t *testing.T) {
	wantToken := "secret-key"
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") == wantToken {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer s.Close()

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}

	passBody := false
	handler := MakeExternalAuthHandler(next, time.Second*5, s.URL, passBody)

	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	req.Header.Set("X-Token", wantToken)

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
	wantSubstring := "context deadline exceeded\n"
	if !strings.HasSuffix(string(rr.Body.Bytes()), wantSubstring) {
		t.Errorf("Body incorrect, want to have suffix: %q, but got %q", []byte(wantSubstring), rr.Body)
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
