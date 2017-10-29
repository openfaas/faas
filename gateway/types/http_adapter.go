// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"fmt"
	"net/http"
)

// WriteAdapter adapts a ResponseWriter
type WriteAdapter struct {
	Writer     http.ResponseWriter
	HttpResult *HttpResult
}
type HttpResult struct {
	HeaderCode int
}

//NewWriteAdapter create a new NewWriteAdapter
func NewWriteAdapter(w http.ResponseWriter) WriteAdapter {
	return WriteAdapter{Writer: w, HttpResult: &HttpResult{}}
}

//Header adapts Header
func (w WriteAdapter) Header() http.Header {
	return w.Writer.Header()
}

// Write adapts Write
func (w WriteAdapter) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}

// WriteHeader adapts WriteHeader
func (w WriteAdapter) WriteHeader(i int) {
	w.Writer.WriteHeader(i)
	w.HttpResult.HeaderCode = i
	fmt.Println("GetHeaderCode before", w.HttpResult.HeaderCode)
}

// GetHeaderCode result from WriteHeader
func (w *WriteAdapter) GetHeaderCode() int {
	return w.HttpResult.HeaderCode
}
