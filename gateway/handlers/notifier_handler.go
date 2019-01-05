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

		writer := newWriteInterceptor(w)
		next(&writer, r)

		url := r.URL.String()
		for _, notifier := range notifiers {
			notifier.Notify(r.Method, url, url, writer.Status(), time.Since(then))
		}
	}
}

func newWriteInterceptor(w http.ResponseWriter) writeInterceptor {
	return writeInterceptor{
		w: w,
	}
}

type writeInterceptor struct {
	CapturedStatusCode int
	w                  http.ResponseWriter
}

func (c *writeInterceptor) Status() int {
	if c.CapturedStatusCode == 0 {
		return http.StatusOK
	}
	return c.CapturedStatusCode
}

func (c *writeInterceptor) Header() http.Header {
	return c.w.Header()
}

func (c *writeInterceptor) Write(data []byte) (int, error) {
	return c.w.Write(data)
}

func (c *writeInterceptor) WriteHeader(code int) {
	c.CapturedStatusCode = code
	c.w.WriteHeader(code)
}
