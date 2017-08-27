// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

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
