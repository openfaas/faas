// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/queue"
)

// MakeQueuedProxy accepts work onto a queue
func MakeQueuedProxy(metrics metrics.MetricOptions, wildcard bool, logger *logrus.Logger, canQueueRequests queue.CanQueueRequests) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		vars := mux.Vars(r)
		name := vars["name"]

		callbackURLHeader := r.Header.Get("X-Callback-Url")
		var callbackURL *url.URL

		if len(callbackURLHeader) > 0 {
			urlVal, urlErr := url.Parse(callbackURLHeader)
			if urlErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(urlErr.Error()))
				return
			}

			callbackURL = urlVal
		}
		req := &queue.Request{
			Function:    name,
			Body:        body,
			Method:      r.Method,
			QueryString: r.URL.RawQuery,
			Header:      r.Header,
			CallbackURL: callbackURL,
		}

		err = canQueueRequests.Queue(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusAccepted)

	}
}
