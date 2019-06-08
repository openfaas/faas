// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas/gateway/handlers"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/plugin"
	"github.com/openfaas/faas/gateway/scaling"
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

	// credentials is used for service-to-service auth
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

	reverseProxy := types.NewHTTPClientReverseProxy(config.FunctionsProviderURL,
		config.UpstreamTimeout,
		config.MaxIdleConns,
		config.MaxIdleConnsPerHost)

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

	var serviceAuthInjector handlers.AuthInjector

	if config.UseBasicAuth {
		serviceAuthInjector = &handlers.BasicAuthInjector{Credentials: credentials}
	}

	decorateExternalAuth := handlers.MakeExternalAuthHandler

	faasHandlers.Proxy = handlers.MakeForwardingProxyHandler(reverseProxy, functionNotifiers, functionURLResolver, functionURLTransformer, nil)

	faasHandlers.ListFunctions = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer, serviceAuthInjector)
	faasHandlers.DeployFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer, serviceAuthInjector)
	faasHandlers.DeleteFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer, serviceAuthInjector)
	faasHandlers.UpdateFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer, serviceAuthInjector)
	faasHandlers.QueryFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer, serviceAuthInjector)
	faasHandlers.InfoHandler = handlers.MakeInfoHandler(handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer, serviceAuthInjector))
	faasHandlers.SecretHandler = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer, serviceAuthInjector)

	alertHandler := plugin.NewExternalServiceQuery(*config.FunctionsProviderURL, serviceAuthInjector)
	faasHandlers.Alert = handlers.MakeNotifierWrapper(
		handlers.MakeAlertHandler(alertHandler),
		forwardingNotifiers,
	)

	if config.UseNATS() {
		log.Println("Async enabled: Using NATS Streaming.")
		maxReconnect := 60
		interval := time.Second * 2

		defaultNATSConfig := natsHandler.NewDefaultNATSConfig(maxReconnect, interval)

		natsQueue, queueErr := natsHandler.CreateNATSQueue(*config.NATSAddress, *config.NATSPort, defaultNATSConfig)
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

	faasHandlers.ScaleFunction = handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer, serviceAuthInjector)

	if credentials != nil {
		faasHandlers.Alert =
			decorateExternalAuth(faasHandlers.Alert, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)
		faasHandlers.UpdateFunction =
			decorateExternalAuth(faasHandlers.UpdateFunction, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)
		faasHandlers.DeleteFunction =
			decorateExternalAuth(faasHandlers.DeleteFunction, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)
		faasHandlers.DeployFunction =
			decorateExternalAuth(faasHandlers.DeployFunction, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)
		faasHandlers.ListFunctions =
			decorateExternalAuth(faasHandlers.ListFunctions, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)
		faasHandlers.ScaleFunction =
			decorateExternalAuth(faasHandlers.ScaleFunction, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)
		faasHandlers.QueryFunction =
			decorateExternalAuth(faasHandlers.QueryFunction, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)
		faasHandlers.InfoHandler =
			decorateExternalAuth(faasHandlers.InfoHandler, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)
		faasHandlers.AsyncReport =
			decorateExternalAuth(faasHandlers.AsyncReport, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)
		faasHandlers.SecretHandler =
			decorateExternalAuth(faasHandlers.SecretHandler, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)
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
		r.PathPrefix("/ui/").Handler(decorateExternalAuth(uiHandler.ServeHTTP, config.UpstreamTimeout, config.AuthProxyURL, config.AuthProxyPassBody)).Methods(http.MethodGet)
	} else {
		r.PathPrefix("/ui/").Handler(uiHandler).Methods(http.MethodGet)
	}

	//Start metrics server in a goroutine
	go runMetricsServer()

	r.HandleFunc("/healthz", handlers.MakeForwardingProxyHandler(reverseProxy, forwardingNotifiers, urlResolver, nilURLTransformer, serviceAuthInjector)).Methods(http.MethodGet)

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

//runMetricsServer Listen on a separate HTTP port for Prometheus metrics to keep this accessible from
// the internal network only.
func runMetricsServer() {
	metricsHandler := metrics.PrometheusHandler()
	router := mux.NewRouter()
	router.Handle("/metrics", metricsHandler)
	router.HandleFunc("/healthz", handlers.HealthzHandler)

	port := 8082
	readTimeout := 5 * time.Second
	writeTimeout := 5 * time.Second

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
		Handler:        router,
	}

	log.Fatal(s.ListenAndServe())
}
