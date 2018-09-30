package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/queue"
)

// Dummy queue for saving only a single element
type SingletonRequestQueue struct {
	request **queue.Request
}

func (q SingletonRequestQueue) Queue(req *queue.Request) error {
	if q.request == nil {
		return errors.New("SingletonRequestQueue is not initialized")
	}
	*q.request = req
	return nil
}

func (q SingletonRequestQueue) GetRequest() (*queue.Request, error) {
	if q.request == nil {
		return nil, errors.New("SingletonRequestQueue is not initialized")
	}
	return *q.request, nil
}

func Test_MakeQueuedProxy(t *testing.T) {
	rr := httptest.NewRecorder()
	srcBytes := []byte("hello world")
	funcName := "testfunc"
	reader := bytes.NewReader(srcBytes)
	url := fmt.Sprintf("/function/%s/?code=1", funcName)
	req, err := http.NewRequest(http.MethodPost, url, reader)
	req.Header.Set("X-Source", "unit-test")
	req.Header.Set("X-Callback-Url", "http://callback")

	if err != nil {
		t.Fatal(err)
	}

	metricsOptions := metrics.BuildMetricsOptions()
	queue := SingletonRequestQueue{request: new(*queue.Request)}

	router := mux.NewRouter()
	functionProxy := func(w http.ResponseWriter, r *http.Request) {
		proxyHandler := MakeQueuedProxy(metricsOptions, true, queue)
		proxyHandler(w, r)
	}
	router.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", functionProxy)
	router.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", functionProxy)
	router.ServeHTTP(rr, req)

	required := http.StatusAccepted
	if status := rr.Code; status != required {
		t.Errorf("handler returned wrong status code - got: %v, want: %v",
			status, required)
		t.Fatal()
	}

	queueReq, err := queue.GetRequest()

	if queueReq == nil {
		t.Errorf("Request in queue should not be nil")
		t.Fail()
	}

	if funcName != queueReq.Function {
		t.Errorf("Function name - want: %s, got: %s", funcName, queueReq.Function)
		t.Fail()
	}

	if req.Method != queueReq.Method {
		t.Errorf("Method - want: %s, got: %s", req.Method, queueReq.Method)
		t.Fail()
	}

	if req.URL.RawQuery != queueReq.QueryString {
		t.Errorf("QueryString - want: %s, got: %s", req.URL.RawQuery, queueReq.QueryString)
		t.Fail()
	}

	if string(srcBytes) != string(queueReq.Body) {
		t.Errorf("Body - want: %v, got: %v", string(srcBytes), string(queueReq.Body))
		t.Fail()
	}

	if req.Host != queueReq.Host {
		t.Errorf("Host - want: %s, got: %s", req.Host, queueReq.Host)
		t.Fail()
	}

	if req.Header.Get("X-Source") != queueReq.Header.Get("X-Source") {
		t.Errorf("Header X-Source - want: %s, got: %s", req.Header.Get("X-Source"), queueReq.Header.Get("X-Source"))
		t.Fail()
	}

	if req.Header.Get("X-Callback-Url") != queueReq.CallbackURL.String() {
		t.Errorf("Callback URL - want: %s, got: %s", req.Header.Get("X-Callback-Url"), queueReq.CallbackURL)
		t.Fail()
	}

}
