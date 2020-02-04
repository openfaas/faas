package handlers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func Test_logsProxyDoesNotLeakGoroutinesWhenProviderClosesConnection(t *testing.T) {
	defer goleak.VerifyNoLeaks(t)

	expectedMsg := "name: funcFoo msg: test message"

	// mock log provider that sends one line and immediately closes the connection
	mockLogsUpstreamEndpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected method '%s' but got '%s'", http.MethodGet, r.Method)
		}

		if r.URL.Path != upstreamLogsEndpoint {
			t.Fatalf("expected path '%s' but got '%s'", upstreamLogsEndpoint, r.URL.Path)
		}

		w.Header().Set(http.CanonicalHeaderKey("Connection"), "Keep-Alive")
		w.Header().Set(http.CanonicalHeaderKey("Transfer-Encoding"), "chunked")
		w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/x-ndjson")
		w.WriteHeader(http.StatusOK)

		msg := fmt.Sprintf("name: %s msg: test message", r.URL.Query().Get("name"))
		_, err := w.Write([]byte(msg))
		if err != nil {
			t.Fatalf("failed to write test log message: %s", err)
		}
	}))
	defer mockLogsUpstreamEndpoint.Close()

	logProviderURL, _ := url.Parse(mockLogsUpstreamEndpoint.URL)

	logHandler := NewLogHandlerFunc(*logProviderURL, time.Minute)
	testSrv := httptest.NewServer(http.HandlerFunc(logHandler))
	defer testSrv.Close()

	resp, err := http.Get(testSrv.URL + "?name=funcFoo")
	if err != nil {
		t.Fatalf("unexpected error sending log request: %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error reading the response body: %s", err)
	}

	if string(body) != string(expectedMsg) {
		t.Fatalf("expected log message %s, got: %s", expectedMsg, body)
	}
}

func Test_logsProxyDoesNotLeakGoroutinesWhenClientClosesConnection(t *testing.T) {
	defer goleak.VerifyNoLeaks(t)

	// mock log provider that sends one line and holds until we cancel the context
	mockLogsUpstreamEndpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cn, ok := w.(http.CloseNotifier)
		if !ok {
			http.Error(w, "cannot stream", http.StatusInternalServerError)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "cannot stream", http.StatusInternalServerError)
			return
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected method '%s' but got '%s'", http.MethodGet, r.Method)
		}

		if r.URL.Path != upstreamLogsEndpoint {
			t.Fatalf("expected path '%s' but got '%s'", upstreamLogsEndpoint, r.URL.Path)
		}

		w.Header().Set(http.CanonicalHeaderKey("Connection"), "Keep-Alive")
		w.Header().Set(http.CanonicalHeaderKey("Transfer-Encoding"), "chunked")
		w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/x-ndjson")
		w.WriteHeader(http.StatusOK)

		msg := fmt.Sprintf("name: %s msg: test message", r.URL.Query().Get("name"))
		_, err := w.Write([]byte(msg))
		if err != nil {
			t.Fatalf("failed to write test log message: %s", err)
		}

		flusher.Flush()

		// "wait for connection to close"
		<-cn.CloseNotify()

	}))
	defer mockLogsUpstreamEndpoint.Close()

	logProviderURL, _ := url.Parse(mockLogsUpstreamEndpoint.URL)

	logHandler := NewLogHandlerFunc(*logProviderURL, time.Minute)
	testSrv := httptest.NewServer(http.HandlerFunc(logHandler))
	defer testSrv.Close()

	reqContext, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequest(http.MethodGet, testSrv.URL+"?name=funcFoo", nil)

	req = req.WithContext(reqContext)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error sending log request: %s", err)
	}

	errCh := make(chan error, 1)
	go func() {
		defer resp.Body.Close()
		defer close(errCh)
		_, err := ioutil.ReadAll(resp.Body)
		errCh <- err
	}()
	cancel()
	err = <-errCh
	if err != context.Canceled {
		t.Fatalf("unexpected error reading the response body: %s", err)
	}

}
