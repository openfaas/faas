// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.

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
