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
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/queue"
	"github.com/openfaas/faas/gateway/scaling"
)

// MakeQueuedProxy accepts work onto a queue
func MakeQueuedProxy(metrics metrics.MetricOptions, wildcard bool, queuer queue.RequestQueuer, pathTransformer URLPathTransformer, defaultNS string, functionCacher scaling.FunctionCacher, serviceQuery scaling.ServiceQuery) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		callbackURL, err := getCallbackURLHeader(r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		name := vars["name"]

		queueName, err := getQueueName(name, functionCacher, serviceQuery)

		req := &queue.Request{
			Function:    name,
			Body:        body,
			Method:      r.Method,
			QueryString: r.URL.RawQuery,
			Path:        pathTransformer.Transform(r),
			Header:      r.Header,
			Host:        r.Host,
			CallbackURL: callbackURL,
			QueueName:   queueName,
		}

		if len(queueName) > 0 {
			log.Printf("Queueing %s to: %s\n", name, queueName)
		}

		if err = queuer.Queue(req); err != nil {
			fmt.Printf("Queue error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func getQueueName(name string, cache scaling.FunctionCacher, serviceQuery scaling.ServiceQuery) (queueName string, err error) {
	fn, ns := getNameParts(name)

	query, hit := cache.Get(fn, ns)
	if !hit {
		queryResponse, err := serviceQuery.GetReplicas(fn, ns)
		if err != nil {
			return "", err
		}
		cache.Set(fn, ns, queryResponse)
	}

	query, _ = cache.Get(fn, ns)

	queueName = ""
	if query.Annotations != nil {
		if v := (*query.Annotations)["com.openfaas.queue"]; len(v) > 0 {
			queueName = v
		}
	}
	return queueName, err
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
