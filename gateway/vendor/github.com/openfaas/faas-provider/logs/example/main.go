package main

import (
	"context"
	"net/http"
	"time"

	"github.com/openfaas/faas-provider/logs"
)

// staticLogRequestor implements the logs Requestor returning a static stream of logs
type staticLogRequestor struct {
	logs []string
}

func (s staticLogRequestor) Query(ctx context.Context, r logs.Request) (<-chan logs.Message, error) {
	resp := make(chan logs.Message, len(s.logs))

	// A real implementation would possibly run a query to their log storage here, if it returns a
	// channel, and pass that channel to the go routine below instead of ranging over `s.logs`
	// If the log storage backend client does not return a channel, the query would need to
	// occur at the beginning of the goroutine below ... or a separate goroutine

	go func() {
		for _, m := range s.logs {
			// always watch the ctx to timeout/cancel/finish
			if ctx.Err() != nil {
				return
			}

			resp <- logs.Message{
				Name:      r.Name,
				Instance:  "fake",
				Timestamp: time.Now(),
				Text:      m,
			}
		}
	}()

	return resp, nil
}

func main() {
	requestor := staticLogRequestor{logs: []string{"msg1", "msg2", "something interesting"}}
	http.HandleFunc("/system/logs", logs.NewLogHandlerFunc(requestor, 10*time.Second))
	http.ListenAndServe(":80", nil)
}
