// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"net/http"
	"time"
)

// MakeNotifierWrapper wraps a http.HandlerFunc in an interceptor to pass to HTTPNotifier
func MakeNotifierWrapper(next http.HandlerFunc, notifiers []HTTPNotifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		then := time.Now()

		writer := newCustomWriter(w)
		next(w, r)

		url := r.URL.String()
		for _, notifier := range notifiers {
			notifier.Notify(r.Method, url, url, writer.CapturedStatusCode, time.Since(then))
		}
	}
}

func newCustomWriter(w http.ResponseWriter) customWriter {
	return customWriter{
		w: w,
	}
}

type customWriter struct {
	CapturedStatusCode int
	w                  http.ResponseWriter
}

func (c *customWriter) Write(data []byte) (int, error) {
	return c.w.Write(data)
}

func (c *customWriter) WriteHeader(code int) {
	c.CapturedStatusCode = code
	c.w.WriteHeader(code)
}
