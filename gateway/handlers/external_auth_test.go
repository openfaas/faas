package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
	handler := MakeExternalAuthHandler(next, s.URL, passBody)

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
	handler := MakeExternalAuthHandler(next, s.URL, passBody)

	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	rr := httptest.NewRecorder()
	handler(rr, req)
	want := http.StatusNotImplemented
	if rr.Code != want {
		t.Errorf("Status incorrect, want: %d, but got %d", want, rr.Code)
	}
}

func MakeExternalAuthHandler(next http.HandlerFunc, upstreamURL string, passBody bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest(http.MethodGet, upstreamURL, nil)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if res.Body != nil {
			defer res.Body.Close()
		}

		if res.StatusCode == http.StatusOK {
			next.ServeHTTP(w, r)
		}
		w.WriteHeader(res.StatusCode)
	}
}
