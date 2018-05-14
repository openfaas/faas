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
		log.Fatalln("You must provide an external provider via 'functions_provider_url' env-var.")
	}

	log.Printf("Binding to external function provider: %s", config.FunctionsProviderURL)

	var credentials *types.BasicAuthCredentials

	if config.UseBasicAuth {
		var readErr error
		reader := types.ReadBasicAuthFromDisk{
			SecretMountPath: config.SecretMountPath,
		}
		credentials, readErr = reader.Read()

		if readErr != nil {
			log.Panicf(readErr.Error())
		}
	}

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
		natsQueue, queueErr := natsHandler.CreateNatsQueue(*config.NATSAddress, *config.NATSPort, natsHandler.DefaultNatsConfig{})
		if queueErr != nil {
			log.Fatalln(queueErr)
		}

		faasHandlers.QueuedProxy = handlers.MakeQueuedProxy(metricsOptions, true, natsQueue)
		faasHandlers.AsyncReport = handlers.MakeAsyncReport(metricsOptions)
	}

	prometheusQuery := metrics.NewPrometheusQuery(config.PrometheusHost, config.PrometheusPort, &http.Client{})
	faasHandlers.ListFunctions = metrics.AddMetricsHandler(faasHandlers.ListFunctions, prometheusQuery)
	faasHandlers.Proxy = handlers.MakeCallIDMiddleware(faasHandlers.Proxy)

	if credentials != nil {
		faasHandlers.UpdateFunction =
			handlers.DecorateWithBasicAuth(faasHandlers.UpdateFunction, credentials)
		faasHandlers.DeleteFunction =
			handlers.DecorateWithBasicAuth(faasHandlers.DeleteFunction, credentials)
		faasHandlers.DeployFunction =
			handlers.DecorateWithBasicAuth(faasHandlers.DeployFunction, credentials)
		faasHandlers.ListFunctions =
			handlers.DecorateWithBasicAuth(faasHandlers.ListFunctions, credentials)
	}

	r := mux.NewRouter()
	// max wait time to start a function = maxPollCount * functionPollInterval

	scalingConfig := handlers.ScalingConfig{
		MaxPollCount:         uint(1000),
		FunctionPollInterval: time.Millisecond * 10,
		CacheExpiry:          time.Second * 5, // freshness of replica values before going stale
	}

	scalingProxy := handlers.MakeScalingHandler(faasHandlers.Proxy, queryFunction, scalingConfig)
	// r.StrictSlash(false)	// This didn't work, so register routes twice.
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", scalingProxy)
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", scalingProxy)

	r.HandleFunc("/system/info", handlers.MakeInfoHandler(handlers.MakeForwardingProxyHandler(
		reverseProxy, forwardingNotifiers, urlResolver))).Methods(http.MethodGet)

	r.HandleFunc("/system/alert", faasHandlers.Alert)

	r.HandleFunc("/system/function/{name:[-a-zA-Z_0-9]+}", queryFunction).Methods(http.MethodGet)
	r.HandleFunc("/system/functions", faasHandlers.ListFunctions).Methods(http.MethodGet)
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

	uiHandler := http.StripPrefix("/ui", fsCORS)
	if credentials != nil {
		r.PathPrefix("/ui/").Handler(handlers.DecorateWithBasicAuth(uiHandler.ServeHTTP, credentials)).Methods(http.MethodGet)
	} else {
		r.PathPrefix("/ui/").Handler(uiHandler).Methods(http.MethodGet)
	}

	metricsHandler := metrics.PrometheusHandler()
	r.Handle("/metrics", metricsHandler)
	r.HandleFunc("/healthz", handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver)).Methods(http.MethodGet)

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
