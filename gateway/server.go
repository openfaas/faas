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
	"github.com/openfaas/faas/gateway/scaling"

	"github.com/openfaas/faas-provider/auth"
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

	var credentials *auth.BasicAuthCredentials

	if config.UseBasicAuth {
		var readErr error
		reader := auth.ReadBasicAuthFromDisk{
			SecretMountPath: config.SecretMountPath,
		}
		credentials, readErr = reader.Read()

		if readErr != nil {
			log.Panicf(readErr.Error())
		}
	}

	var faasHandlers types.HandlerSet

	servicePollInterval := time.Second * 5

	metricsOptions := metrics.BuildMetricsOptions()
	exporter := metrics.NewExporter(metricsOptions, credentials)
	exporter.StartServiceWatcher(*config.FunctionsProviderURL, metricsOptions, "func", servicePollInterval)
	metrics.RegisterExporter(exporter)

	reverseProxy := types.NewHTTPClientReverseProxy(config.FunctionsProviderURL, config.UpstreamTimeout)

	loggingNotifier := handlers.LoggingNotifier{}
	prometheusNotifier := handlers.PrometheusFunctionNotifier{
		Metrics: &metricsOptions,
	}
	prometheusServiceNotifier := handlers.PrometheusServiceNotifier{
		ServiceMetrics: metricsOptions.ServiceMetrics,
	}

	functionNotifiers := []handlers.HTTPNotifier{loggingNotifier, prometheusNotifier}
	forwardingNotifiers := []handlers.HTTPNotifier{loggingNotifier, prometheusServiceNotifier}

	urlResolver := handlers.SingleHostBaseURLResolver{BaseURL: config.FunctionsProviderURL.String()}
	var functionURLResolver handlers.BaseURLResolver
	var functionURLTransformer handlers.URLPathTransformer
	nilURLTransformer := handlers.TransparentURLPathTransformer{}

	if config.DirectFunctions {
		functionURLResolver = handlers.FunctionAsHostBaseURLResolver{FunctionSuffix: config.DirectFunctionsSuffix}
		functionURLTransformer = handlers.FunctionPrefixTrimmingURLPathTransformer{}
	} else {
		functionURLResolver = urlResolver
		functionURLTransformer = nilURLTransformer
	}

	faasHandlers.Proxy = handlers.MakeForwardingProxyHandler(reverseProxy, functionNotifiers, functionURLResolver, functionURLTransformer)

	faasHandlers.RoutelessProxy = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer)
	faasHandlers.ListFunctions = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer)
	faasHandlers.DeployFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer)
	faasHandlers.DeleteFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer)
	faasHandlers.UpdateFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer)
	faasHandlers.QueryFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer)
	faasHandlers.InfoHandler = handlers.MakeInfoHandler(handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer))
	faasHandlers.SecretHandler = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer)

	alertHandler := plugin.NewExternalServiceQuery(*config.FunctionsProviderURL, credentials)
	faasHandlers.Alert = handlers.MakeNotifierWrapper(
		handlers.MakeAlertHandler(alertHandler),
		forwardingNotifiers,
	)

	if config.UseNATS() {
		log.Println("Async enabled: Using NATS Streaming.")
		natsQueue, queueErr := natsHandler.CreateNatsQueue(*config.NATSAddress, *config.NATSPort, natsHandler.DefaultNatsConfig{})
		if queueErr != nil {
			log.Fatalln(queueErr)
		}

		faasHandlers.QueuedProxy = handlers.MakeNotifierWrapper(
			handlers.MakeCallIDMiddleware(handlers.MakeQueuedProxy(metricsOptions, true, natsQueue, functionURLTransformer)),
			forwardingNotifiers,
		)

		faasHandlers.AsyncReport = handlers.MakeNotifierWrapper(
			handlers.MakeAsyncReport(metricsOptions),
			forwardingNotifiers,
		)
	}

	prometheusQuery := metrics.NewPrometheusQuery(config.PrometheusHost, config.PrometheusPort, &http.Client{})
	faasHandlers.ListFunctions = metrics.AddMetricsHandler(faasHandlers.ListFunctions, prometheusQuery)
	faasHandlers.Proxy = handlers.MakeCallIDMiddleware(faasHandlers.Proxy)

	faasHandlers.ScaleFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer)

	if credentials != nil {
		faasHandlers.UpdateFunction =
			auth.DecorateWithBasicAuth(faasHandlers.UpdateFunction, credentials)
		faasHandlers.DeleteFunction =
			auth.DecorateWithBasicAuth(faasHandlers.DeleteFunction, credentials)
		faasHandlers.DeployFunction =
			auth.DecorateWithBasicAuth(faasHandlers.DeployFunction, credentials)
		faasHandlers.ListFunctions =
			auth.DecorateWithBasicAuth(faasHandlers.ListFunctions, credentials)
		faasHandlers.ScaleFunction =
			auth.DecorateWithBasicAuth(faasHandlers.ScaleFunction, credentials)
		faasHandlers.QueryFunction =
			auth.DecorateWithBasicAuth(faasHandlers.QueryFunction, credentials)
		faasHandlers.InfoHandler =
			auth.DecorateWithBasicAuth(faasHandlers.InfoHandler, credentials)
		faasHandlers.AsyncReport =
			auth.DecorateWithBasicAuth(faasHandlers.AsyncReport, credentials)
		faasHandlers.SecretHandler =
			auth.DecorateWithBasicAuth(faasHandlers.SecretHandler, credentials)
	}

	r := mux.NewRouter()
	// max wait time to start a function = maxPollCount * functionPollInterval

	functionProxy := faasHandlers.Proxy

	if config.ScaleFromZero {
		scalingConfig := scaling.ScalingConfig{
			MaxPollCount:         uint(1000),
			SetScaleRetries:      uint(20),
			FunctionPollInterval: time.Millisecond * 50,
			CacheExpiry:          time.Second * 5, // freshness of replica values before going stale
			ServiceQuery:         alertHandler,
		}

		functionProxy = handlers.MakeScalingHandler(faasHandlers.Proxy, scalingConfig)
	}
	// r.StrictSlash(false)	// This didn't work, so register routes twice.
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", functionProxy)
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", functionProxy)
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/{params:.*}", functionProxy)

	r.HandleFunc("/system/info", faasHandlers.InfoHandler).Methods(http.MethodGet)
	r.HandleFunc("/system/alert", faasHandlers.Alert).Methods(http.MethodPost)

	r.HandleFunc("/system/function/{name:[-a-zA-Z_0-9]+}", faasHandlers.QueryFunction).Methods(http.MethodGet)
	r.HandleFunc("/system/functions", faasHandlers.ListFunctions).Methods(http.MethodGet)
	r.HandleFunc("/system/functions", faasHandlers.DeployFunction).Methods(http.MethodPost)
	r.HandleFunc("/system/functions", faasHandlers.DeleteFunction).Methods(http.MethodDelete)
	r.HandleFunc("/system/functions", faasHandlers.UpdateFunction).Methods(http.MethodPut)
	r.HandleFunc("/system/scale-function/{name:[-a-zA-Z_0-9]+}", faasHandlers.ScaleFunction).Methods(http.MethodPost)

	r.HandleFunc("/system/secrets", faasHandlers.SecretHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)

	if faasHandlers.QueuedProxy != nil {
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}/", faasHandlers.QueuedProxy).Methods(http.MethodPost)
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}", faasHandlers.QueuedProxy).Methods(http.MethodPost)
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}/{params:.*}", faasHandlers.QueuedProxy).Methods(http.MethodPost)

		r.HandleFunc("/system/async-report", handlers.MakeNotifierWrapper(faasHandlers.AsyncReport, forwardingNotifiers))
	}

	fs := http.FileServer(http.Dir("./assets/"))

	// This URL allows access from the UI to the OpenFaaS store
	allowedCORSHost := "raw.githubusercontent.com"
	fsCORS := handlers.DecorateWithCORS(fs, allowedCORSHost)

	uiHandler := http.StripPrefix("/ui", fsCORS)
	if credentials != nil {
		r.PathPrefix("/ui/").Handler(auth.DecorateWithBasicAuth(uiHandler.ServeHTTP, credentials)).Methods(http.MethodGet)
	} else {
		r.PathPrefix("/ui/").Handler(uiHandler).Methods(http.MethodGet)
	}

	metricsHandler := metrics.PrometheusHandler()
	r.Handle("/metrics", metricsHandler)
	r.HandleFunc("/healthz", handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer)).Methods(http.MethodGet)

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
