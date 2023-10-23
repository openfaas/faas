package types

import (
	"net/http"
	"net/url"
)

// Request for asynchronous processing
type QueueRequest struct {
	// Header from HTTP request
	Header http.Header `json:"Header,omitempty"`

	// Host from HTTP request
	Host string `json:"Host,omitempty"`

	// Body from HTTP request to use for invocation
	Body []byte `json:"Body,omitempty"`

	// Method from HTTP request
	Method string `json:"Method"`

	// Path from HTTP request
	Path string `json:"Path,omitempty"`

	// QueryString from HTTP request
	QueryString string `json:"QueryString,omitempty"`

	// Function name to invoke
	Function string `json:"Function"`

	// QueueName to publish the request to, leave blank
	// for default.
	QueueName string `json:"QueueName,omitempty"`

	// Annotations defines a collection of meta-data that can be used by
	// the queue worker when processing the queued request.
	Annotations map[string]string `json:"Annotations,omitempty"`

	// Used by queue worker to submit a result
	CallbackURL *url.URL `json:"CallbackUrl,omitempty"`
}

// RequestQueuer can public a request to be executed asynchronously
type RequestQueuer interface {
	Queue(req *QueueRequest) error
}
