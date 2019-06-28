package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/goleak"
)

var queryTimeout = 30 * time.Second

func Test_logsHandlerDoesNotLeakGoroutinesWhenProviderClosesStream(t *testing.T) {
	defer goleak.VerifyNoLeaks(t)

	msgs := []Message{
		Message{Name: "funcFoo", Text: "msg 0"},
		Message{Name: "funcFoo", Text: "msg 1"},
	}

	var expected bytes.Buffer
	json.NewEncoder(&expected).Encode(msgs[0])
	json.NewEncoder(&expected).Encode(msgs[1])

	querier := newFakeQueryRequester(msgs, nil)
	logHandler := NewLogHandlerFunc(querier, queryTimeout)
	testSrv := httptest.NewServer(http.HandlerFunc(logHandler))
	defer testSrv.Close()

	resp, err := http.Get(testSrv.URL + "?name=funcFoo")
	if err != nil {
		t.Fatalf("unexpected error sending log request: %s", err)
	}

	querier.Close()

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error reading log response: %s", err)
	}

	if string(body) != expected.String() {
		t.Fatalf("expected log message %s, got: %s", expected.String(), body)
	}
}

func Test_logsHandlerDoesNotLeakGoroutinesWhenClientClosesConnection(t *testing.T) {
	defer goleak.VerifyNoLeaks(t)

	msgs := []Message{
		Message{Name: "funcFoo", Text: "msg 0"},
		Message{Name: "funcFoo", Text: "msg 1"},
	}

	querier := newFakeQueryRequester(msgs, nil)
	logHandler := NewLogHandlerFunc(querier, queryTimeout)
	testSrv := httptest.NewServer(http.HandlerFunc(logHandler))
	defer testSrv.Close()

	reqContext, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequest(http.MethodGet, testSrv.URL+"?name=funcFoo", nil)

	req = req.WithContext(reqContext)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error sending log request: %s", err)
	}

	go func() {
		defer resp.Body.Close()
		_, err := ioutil.ReadAll(resp.Body)
		if err != context.Canceled {
			t.Fatalf("unexpected error reading log response: %s", err)
		}
	}()
	cancel()
}

func Test_GETRequestParsing(t *testing.T) {
	sinceTime, _ := time.Parse(time.RFC3339, "2019-02-16T09:10:06+00:00")
	scenarios := []struct {
		name            string
		rawQueryStr     string
		err             string
		expectedRequest Request
	}{
		{
			name:            "empty query creates an empty request",
			rawQueryStr:     "",
			err:             "",
			expectedRequest: Request{},
		},
		{
			name:            "name only query",
			rawQueryStr:     "name=foobar",
			err:             "",
			expectedRequest: Request{Name: "foobar"},
		},
		{
			name:            "name only query",
			rawQueryStr:     "name=foobar",
			err:             "",
			expectedRequest: Request{Name: "foobar"},
		},
		{
			name:            "multiple name values selects the last value",
			rawQueryStr:     "name=foobar&name=theactual name",
			err:             "",
			expectedRequest: Request{Name: "theactual name"},
		},
		{
			name:        "valid request with every parameter",
			rawQueryStr: "name=foobar&since=2019-02-16T09%3A10%3A06%2B00%3A00&tail=5&follow=true",
			err:         "",
			expectedRequest: Request{
				Name:   "foobar",
				Since:  &sinceTime,
				Tail:   5,
				Follow: true,
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			req.URL.RawQuery = s.rawQueryStr
			logRequest, err := parseRequest(req)
			equalError(t, s.err, err)

			if logRequest.String() != s.expectedRequest.String() {
				t.Errorf("expected log request: %s, got: %s", s.expectedRequest, logRequest)
			}
		})
	}
}

func equalError(t *testing.T, expected string, actual error) {
	if expected == "" && actual == nil {
		return
	}

	if expected == "" && actual != nil {
		t.Errorf("unexpected error: %s", actual.Error())
		return
	}

	if actual.Error() != expected {
		t.Errorf("expected error: %s got: %s", expected, actual.Error())
	}
}

type fakeQueryRequester struct {
	Logs   []Message
	err    error
	stream chan Message
}

func (r fakeQueryRequester) Close() {
	close(r.stream)
	r.stream = nil
}

func (r fakeQueryRequester) Query(context.Context, Request) (<-chan Message, error) {
	if r.err != nil {
		return nil, r.err
	}

	for _, m := range r.Logs {
		r.stream <- m
	}

	return r.stream, nil
}

func newFakeQueryRequester(l []Message, err error) fakeQueryRequester {
	return fakeQueryRequester{
		Logs:   l,
		err:    err,
		stream: make(chan Message, len(l)),
	}

}
