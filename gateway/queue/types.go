// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package queue

import (
	"net/http"
	"net/url"
)

// Request for asynchronous processing
type Request struct {
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
	Queue(req *Request) error
}
