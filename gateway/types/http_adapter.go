// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"log"
	"net/http"
)

// WriteAdapter adapts a ResponseWriter
type WriteAdapter struct {
	Writer     http.ResponseWriter
	HTTPResult *HTTPResult
}

// HTTPResult captures data from forwarded HTTP call
type HTTPResult struct {
	HeaderCode int // HeaderCode is the result of WriteHeader(int)
}

//NewWriteAdapter create a new NewWriteAdapter
func NewWriteAdapter(w http.ResponseWriter) WriteAdapter {
	return WriteAdapter{Writer: w, HTTPResult: &HTTPResult{}}
}

//Header adapts Header
func (w WriteAdapter) Header() http.Header {
	return w.Writer.Header()
}

// Write adapts Write for a straight pass-through
func (w WriteAdapter) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}

// WriteHeader adapts WriteHeader
func (w WriteAdapter) WriteHeader(statusCode int) {
	w.Writer.WriteHeader(statusCode)
	w.HTTPResult.HeaderCode = statusCode

	log.Printf("GetHeaderCode %d", w.HTTPResult.HeaderCode)
}

// GetHeaderCode result from WriteHeader
func (w *WriteAdapter) GetHeaderCode() int {
	return w.HTTPResult.HeaderCode
}
