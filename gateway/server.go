// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas/gateway/handlers"

	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/plugin"
	"github.com/openfaas/faas/gateway/types"
	natsHandler "github.com/openfaas/nats-queue-worker/handler"
)

func main() {

	osEnv := types.OsEnv{}
	readConfig := types.ReadConfig{}
	config := readConfig.Read(osEnv)

	log.Printf("HTTP Read Timeout: %s", config.ReadTimeout)
	log.Printf("HTTP Write Timeout: %s", config.WriteTimeout)

	if !config.UseExternalProvider() {
		log.Fatalln("As of this version of OpenFaaS, you must use external provider even for Docker Swarm.")
	}

	log.Printf("Binding to external function provider: %s", config.FunctionsProviderURL)

	metricsOptions := metrics.BuildMetricsOptions()
	metrics.RegisterMetrics(metricsOptions)

	var faasHandlers types.HandlerSet

	servicePollInterval := time.Second * 5

	reverseProxy := types.NewHTTPClientReverseProxy(config.FunctionsProviderURL, config.UpstreamTimeout)

	loggingNotifier := handlers.LoggingNotifier{}
	prometheusNotifier := handlers.PrometheusFunctionNotifier{
		Metrics: &metricsOptions,
	}
	functionNotifiers := []handlers.HTTPNotifier{loggingNotifier, prometheusNotifier}
	forwardingNotifiers := []handlers.HTTPNotifier{loggingNotifier}

	urlResolver := handlers.SingleHostBaseURLResolver{BaseURL: config.FunctionsProviderURL.String()}
	var functionURLResolver handlers.BaseURLResolver

	if config.DirectFunctions {
		functionURLResolver = handlers.FunctionAsHostBaseURLResolver{FunctionSuffix: config.DirectFunctionsSuffix}
	} else {
		functionURLResolver = urlResolver
	}

	faasHandlers.Proxy = handlers.MakeForwardingProxyHandler(reverseProxy, functionNotifiers, functionURLResolver)

	faasHandlers.RoutelessProxy = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver)
	faasHandlers.ListFunctions = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver)
	faasHandlers.DeployFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver)
	faasHandlers.DeleteFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver)
	faasHandlers.UpdateFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver)
	queryFunction := handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver)

	alertHandler := plugin.NewExternalServiceQuery(*config.FunctionsProviderURL)
	faasHandlers.Alert = handlers.MakeAlertHandler(alertHandler)

	metrics.AttachExternalWatcher(*config.FunctionsProviderURL, metricsOptions, "func", servicePollInterval)

	if config.UseNATS() {
		log.Println("Async enabled: Using NATS Streaming.")
		natsQueue, queueErr := natsHandler.CreateNatsQueue(*config.NATSAddress, *config.NATSPort)
		if queueErr != nil {
			log.Fatalln(queueErr)
		}

		faasHandlers.QueuedProxy = handlers.MakeQueuedProxy(metricsOptions, true, natsQueue)
		faasHandlers.AsyncReport = handlers.MakeAsyncReport(metricsOptions)
	}

	prometheusQuery := metrics.NewPrometheusQuery(config.PrometheusHost, config.PrometheusPort, &http.Client{})
	listFunctions := metrics.AddMetricsHandler(faasHandlers.ListFunctions, prometheusQuery)
	faasHandlers.Proxy = handlers.MakeCallIDMiddleware(faasHandlers.Proxy)
	r := mux.NewRouter()

	// r.StrictSlash(false)	// This didn't work, so register routes twice.
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", faasHandlers.Proxy)
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", faasHandlers.Proxy)

	r.HandleFunc("/system/alert", faasHandlers.Alert)

	r.HandleFunc("/system/function/{name:[-a-zA-Z_0-9]+}", queryFunction).Methods(http.MethodGet)
	r.HandleFunc("/system/functions", listFunctions).Methods(http.MethodGet)
	r.HandleFunc("/system/functions", faasHandlers.DeployFunction).Methods(http.MethodPost)
	r.HandleFunc("/system/functions", faasHandlers.DeleteFunction).Methods(http.MethodDelete)
	r.HandleFunc("/system/functions", faasHandlers.UpdateFunction).Methods(http.MethodPut)

	if faasHandlers.QueuedProxy != nil {
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}/", faasHandlers.QueuedProxy).Methods(http.MethodPost)
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}", faasHandlers.QueuedProxy).Methods(http.MethodPost)

		r.HandleFunc("/system/async-report", faasHandlers.AsyncReport)
	}

	fs := http.FileServer(http.Dir("./assets/"))

	// This URL allows access from the UI to the OpenFaaS store
	allowedCORSHost := "raw.githubusercontent.com"
	fsCORS := handlers.DecorateWithCORS(fs, allowedCORSHost)

	r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui", fsCORS)).Methods(http.MethodGet)

	metricsHandler := metrics.PrometheusHandler()
	r.Handle("/metrics", metricsHandler)
	r.Handle("/", http.RedirectHandler("/ui/", http.StatusMovedPermanently)).Methods(http.MethodGet)

	tcpPort := 8080

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", tcpPort),
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes, // 1MB - can be overridden by setting Server.MaxHeaderBytes.
		Handler:        r,
	}

	log.Fatal(s.ListenAndServe())
}
