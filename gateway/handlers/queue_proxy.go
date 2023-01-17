// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	ftypes "github.com/openfaas/faas-provider/types"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/pkg/middleware"

	"github.com/openfaas/faas/gateway/scaling"
)

// MakeQueuedProxy accepts work onto a queue
func MakeQueuedProxy(metrics metrics.MetricOptions, queuer ftypes.RequestQueuer, pathTransformer middleware.URLPathTransformer, defaultNS string, functionQuery scaling.FunctionQuery) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		if r.Body != nil {
			defer r.Body.Close()

			var err error
			body, err = ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		callbackURL, err := getCallbackURLHeader(r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		name := vars["name"]

		req := &ftypes.QueueRequest{
			Function:    name,
			Body:        body,
			Method:      r.Method,
			QueryString: r.URL.RawQuery,
			Path:        pathTransformer.Transform(r),
			Header:      r.Header,
			Host:        r.Host,
			CallbackURL: callbackURL,
		}

		if err = queuer.Queue(req); err != nil {
			log.Printf("Error queuing request: %v", err)
			http.Error(w, fmt.Sprintf("Error queuing request: %s", err.Error()),
				http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func getCallbackURLHeader(header http.Header) (*url.URL, error) {
	value := header.Get("X-Callback-Url")
	var callbackURL *url.URL

	if len(value) > 0 {
		urlVal, err := url.Parse(value)
		if err != nil {
			return callbackURL, err
		}

		callbackURL = urlVal
	}

	return callbackURL, nil
}

func getNameParts(name string) (fn, ns string) {
	fn = name
	ns = ""

	if index := strings.LastIndex(name, "."); index > 0 {
		fn = name[:index]
		ns = name[index+1:]
	}
	return fn, ns
}
