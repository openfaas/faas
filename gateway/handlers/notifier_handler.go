// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"net/http"
	"time"

	"github.com/openfaas/faas-provider/httputil"
)

// MakeNotifierWrapper wraps a http.HandlerFunc in an interceptor to pass to HTTPNotifier
func MakeNotifierWrapper(next http.HandlerFunc, notifiers []HTTPNotifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		then := time.Now()
		url := r.URL.String()

		writer := httputil.NewHttpWriteInterceptor(w)
		next(writer, r)

		for _, notifier := range notifiers {
			notifier.Notify(r.Method, url, url, writer.Status(), "completed", time.Since(then))
		}
	}
}
