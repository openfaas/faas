package logs

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/openfaas/faas-provider/httputil"
)

// Requester submits queries the logging system.
// This will be passed to the log handler constructor.
type Requester interface {
	// Query submits a log request to the actual logging system.
	Query(context.Context, Request) (<-chan Message, error)
}

// NewLogHandlerFunc creates an http HandlerFunc from the supplied log Requestor.
func NewLogHandlerFunc(requestor Requester, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		cn, ok := w.(http.CloseNotifier)
		if !ok {
			log.Println("LogHandler: response is not a CloseNotifier, required for streaming response")
			http.NotFound(w, r)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			log.Println("LogHandler: response is not a Flusher, required for streaming response")
			http.NotFound(w, r)
			return
		}

		logRequest, err := parseRequest(r)
		if err != nil {
			log.Printf("LogHandler: could not parse request %s", err)
			httputil.Errorf(w, http.StatusUnprocessableEntity, "could not parse the log request")
			return
		}

		ctx, cancelQuery := context.WithTimeout(r.Context(), timeout)
		defer cancelQuery()
		messages, err := requestor.Query(ctx, logRequest)
		if err != nil {
			// add smarter error handling here
			httputil.Errorf(w, http.StatusInternalServerError, "function log request failed")
			return
		}

		// Send the initial headers saying we're gonna stream the response.
		w.Header().Set("Connection", "Keep-Alive")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		// ensure that we always try to send the closing chunk, not the inverted order due to how
		// the defer stack works. We need two flush statements to ensure that the empty slice is
		// sent as its own chunk
		defer flusher.Flush()
		defer w.Write([]byte{})
		defer flusher.Flush()

		jsonEncoder := json.NewEncoder(w)
		for messages != nil {
			select {
			case <-cn.CloseNotify():
				log.Println("LogHandler: client stopped listening")
				return
			case msg, ok := <-messages:
				if !ok {
					log.Println("LogHandler: end of log stream")
					messages = nil
					return
				}

				// serialize and write the msg to the http ResponseWriter
				err := jsonEncoder.Encode(msg)
				if err != nil {
					// can't actually write the status header here so we should json serialize an error
					// and return that because we have already sent the content type and status code
					log.Printf("LogHandler: failed to serialize log message: '%s'\n", msg.String())
					log.Println(err.Error())
					// write json error message here ?
					jsonEncoder.Encode(Message{Text: "failed to serialize log message"})
					flusher.Flush()
					return
				}

				flusher.Flush()
			}
		}

		return
	}
}

// parseRequest extracts the logRequest from the GET variables or from the POST body
func parseRequest(r *http.Request) (logRequest Request, err error) {
	query := r.URL.Query()
	logRequest.Name = getValue(query, "name")
	logRequest.Instance = getValue(query, "instance")
	tailStr := getValue(query, "tail")
	if tailStr != "" {
		logRequest.Tail, err = strconv.Atoi(tailStr)
		if err != nil {
			return logRequest, err
		}
	}
	// ignore error because it will default to false if we can't parse it
	logRequest.Follow, _ = strconv.ParseBool(getValue(query, "follow"))

	sinceStr := getValue(query, "since")
	if sinceStr != "" {
		since, err := time.Parse(time.RFC3339, sinceStr)
		logRequest.Since = &since
		if err != nil {
			return logRequest, err
		}
	}

	return logRequest, nil
}

// getValue returns the value for the given key. If the key has more than one value, it returns the
// last value. if the value does not exist, it returns the empty string.
func getValue(queryValues url.Values, name string) string {
	values := queryValues[name]
	if len(values) == 0 {
		return ""
	}

	return values[len(values)-1]
}
