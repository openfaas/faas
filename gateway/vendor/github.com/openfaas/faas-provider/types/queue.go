package types

import (
	"net/http"
	"net/url"
)

// Request for asynchronous processing
type QueueRequest struct {
	// Header from HTTP request
	Header http.Header

	// Host from HTTP request
	Host string

	// Body from HTTP request to use for invocation
	Body []byte

	// Method from HTTP request
	Method string

	// Path from HTTP request
	Path string

	// QueryString from HTTP request
	QueryString string

	// Function name to invoke
	Function string

	// QueueName to publish the request to, leave blank
	// for default.
	QueueName string

	// Used by queue worker to submit a result
	CallbackURL *url.URL `json:"CallbackUrl"`
}

// RequestQueuer can public a request to be executed asynchronously
type RequestQueuer interface {
	Queue(req *QueueRequest) error
}
