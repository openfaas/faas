package queue

import "net/url"
import "net/http"

// Request for asynchronous processing
type Request struct {
	Header      http.Header
	Body        []byte
	Method      string
	QueryString string
	Function    string
	CallbackURL *url.URL `json:"CallbackUrl"`
}

// CanQueueRequests can take on asynchronous requests
type CanQueueRequests interface {
	Queue(req *Request) error
}
