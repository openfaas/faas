package metrics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testFetchQuery struct {
	res *VectorQueryResponse
	err error
}

func (f testFetchQuery) Fetch(query string) (*VectorQueryResponse, error) {
	return f.res, f.err
}

func TestAddMetricsHandlerServeHTTPReturnsContentTypeOfPassedInHandlerWhenFetchSucceeds(t *testing.T) {
	contentTypes := []string{
		"text/html",
		"application/json",
	}

	for _, contentType := range contentTypes {
		fetcher := testFetchQuery{
			res: &VectorQueryResponse{},
			err: nil,
		}

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", contentType)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		}

		metricsHandler := AddMetricsHandler(handler, fetcher)

		w := httptest.NewRecorder()
		r := &http.Request{}
		metricsHandler.ServeHTTP(w, r)

		expected := contentType
		if contentType := w.Header().Get("Content-Type"); contentType != expected {
			t.Errorf("content type header does not match: got %v want %v",
				contentType, expected)
		}
	}
}

func TestAddMetricsHandlerServeHTTPReturnsContentTypeOfPassedInHandlerWhenFetchFails(t *testing.T) {
	contentTypes := []string{
		"text/html",
		"application/json",
	}

	for _, contentType := range contentTypes {
		fetcher := testFetchQuery{
			res: nil,
			err: errors.New("error"),
		}

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", contentType)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		}

		metricsHandler := AddMetricsHandler(handler, fetcher)

		w := httptest.NewRecorder()
		r := &http.Request{}
		metricsHandler.ServeHTTP(w, r)

		expected := contentType
		if contentType := w.Header().Get("Content-Type"); contentType != expected {
			t.Errorf("content type header does not match: got %v want %v",
				contentType, expected)
		}
	}
}
