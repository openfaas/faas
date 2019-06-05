package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_makeLogger_CopiesResponseHeaders(t *testing.T) {
	handler := http.HandlerFunc(makeLogger(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Unit-Test", "true")
		})))

	s := httptest.NewServer(handler)
	defer s.Close()

	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("X-Unit-Test")
	want := "true"
	if want != got {
		t.Errorf("Header X-Unit-Test, want: %s, got %s", want, got)
	}

}
