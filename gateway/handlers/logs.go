package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const crlf = "\r\n"
const upstreamLogsEndpoint = "/system/logs"

// NewLogHandlerFunc creates and http HandlerFunc from the supplied log Requestor.
func NewLogHandlerFunc(logProvider url.URL, timeout time.Duration) http.HandlerFunc {
	writeRequestURI := false
	if _, exists := os.LookupEnv("write_request_uri"); exists {
		writeRequestURI = exists
	}

	upstreamLogProviderBase := strings.TrimSuffix(logProvider.String(), "/")

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelQuery := context.WithTimeout(r.Context(), timeout)
		defer cancelQuery()

		if r.Body != nil {
			defer r.Body.Close()
		}

		logRequest := buildUpstreamRequest(r, upstreamLogProviderBase, upstreamLogsEndpoint)
		if logRequest.Body != nil {
			defer logRequest.Body.Close()
		}

		cn, ok := w.(http.CloseNotifier)
		if !ok {
			log.Println("LogHandler: response is not a CloseNotifier, required for streaming response")
			http.NotFound(w, r)
			return
		}

		wf, ok := w.(writerFlusher)
		if !ok {
			log.Println("LogHandler: response is not a Flusher, required for streaming response")
			http.NotFound(w, r)
			return
		}

		if writeRequestURI {
			log.Printf("LogProxy: proxying request to %s %s\n", logRequest.Host, logRequest.URL.String())
		}

		ctx, cancel := context.WithCancel(ctx)
		logRequest = logRequest.WithContext(ctx)
		defer cancel()

		logResp, err := http.DefaultTransport.RoundTrip(logRequest)
		if err != nil {
			log.Printf("LogProxy: forwarding request failed: %s\n", err.Error())
			http.Error(w, "log request failed", http.StatusInternalServerError)
			return
		}
		defer logResp.Body.Close()

		switch logResp.StatusCode {
		case http.StatusNotFound, http.StatusNotImplemented:
			w.WriteHeader(http.StatusNotImplemented)
			return
		case http.StatusOK:
			// watch for connection closures and stream data
			// connections and contexts should have cancel methods deferred already
			select {
			case err := <-copyNotify(&unbufferedWriter{wf}, logResp.Body):
				if err != nil {
					log.Printf("LogProxy: error while copy: %s", err.Error())
					return
				}
			case <-cn.CloseNotify():
				log.Printf("LogProxy: client connection closed")
				return
			}
		default:
			http.Error(w, fmt.Sprintf("unknown log request error (%v)", logResp.StatusCode), http.StatusInternalServerError)
		}

		return
	}
}

type writerFlusher interface {
	io.Writer
	http.Flusher
}

// unbufferedWriter is an io Writer that immediately flushes the after every call to Write.
// This can wrap any http.ResponseWriter that also implements Flusher.  This ensures that log
// lines are immediately sent to the client
type unbufferedWriter struct {
	dst writerFlusher
}

// Write writes to the dst writer and then immediately flushes the writer
func (u *unbufferedWriter) Write(p []byte) (n int, err error) {
	n, err = u.dst.Write(p)
	u.dst.Flush()

	return n, err
}

func copyNotify(destination io.Writer, source io.Reader) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(destination, source)
		done <- err
	}()
	return done
}
