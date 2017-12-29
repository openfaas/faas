// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"

	internalHandlers "github.com/openfaas/faas/gateway/handlers"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/metrics/prometheus"
	"github.com/openfaas/faas/gateway/metrics/statsd"
	"github.com/openfaas/faas/gateway/plugin"
	"github.com/openfaas/faas/gateway/replicas"
	"github.com/openfaas/faas/gateway/types"
	natsHandler "github.com/openfaas/nats-queue-worker/handler"
)

func main() {
	log.Println("Ver 3")

	osEnv := types.OsEnv{}
	readConfig := types.ReadConfig{}
	config := readConfig.Read(osEnv)

	logger := setupLogger(config)
	logger.Printf("HTTP Read Timeout: %s", config.ReadTimeout)
	logger.Printf("HTTP Write Timeout: %s", config.WriteTimeout)

	if config.UseExternalProvider() {
		logger.Printf("Binding to external function provider: %s", config.FunctionsProviderURL)
	}

	metrics := setupMetrics(config)

	faasHandlers := setupHandlers(config, logger, metrics)

	setupReplicaWatcher(config, logger, metrics)

	setupNats(config, logger, metrics, &faasHandlers)

	router := setupRouter(&faasHandlers)

	startServer(config, router)
}

func setupMetrics(config types.GatewayConfig) metrics.Metrics {
	if config.StatsDServer != "" {
		return setupStatsDMetrics(config.StatsDServer)
	}
	// setup the prometheus metrics
	return setupPrometheusMetrics()

}

func setupPrometheusMetrics() metrics.Metrics {
	client := prometheus.NewMetrics()

	return client
}

func setupStatsDMetrics(statsDServer string) metrics.Metrics {
	metrics, _ := statsd.NewClient(statsDServer)
	return metrics
}

func setupLogger(config types.GatewayConfig) *logrus.Logger {
	var formatter logrus.Formatter
	if config.LoggerFormat == "JSON" {
		formatter = &logrus.JSONFormatter{}
	} else {
		formatter = &logrus.TextFormatter{}
	}

	logger := logrus.New()
	logger.Formatter = formatter

	// TODO add log rotation
	if config.LoggerFileOutput != "" {
		f, err := os.OpenFile(config.LoggerFileOutput, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			logger.Out = f
		} else {
			logger.Warnf("Unable to open file for output, defaulting to std out: %s", err.Error())
		}
	}

	switch config.LoggerLevel {
	case "DEBUG":
		logger.Level = logrus.DebugLevel
	case "WARNING":
		logger.Level = logrus.WarnLevel
	}

	return logger
}

func setupHandlers(config types.GatewayConfig, logger *logrus.Logger, metrics metrics.Metrics) types.HandlerSet {

	var handlers types.HandlerSet

	if config.UseExternalProvider() {
		handlers = setupExternalProviderHandlers(config, logger, metrics)
	} else {
		handlers = setupSwarmProviderHandlers(config, logger, metrics)
	}

	if config.StatsDServer == "" {
		prometheusQuery := prometheus.NewPrometheusQuery(config.PrometheusHost, config.PrometheusPort, &http.Client{})
		handlers.ListFunctions = prometheus.AddMetricsHandler(handlers.ListFunctions, prometheusQuery)
	}

	return handlers
}

func setupExternalProviderHandlers(config types.GatewayConfig, logger *logrus.Logger, metrics metrics.Metrics) types.HandlerSet {

	var faasHandlers types.HandlerSet

	reverseProxy := httputil.NewSingleHostReverseProxy(config.FunctionsProviderURL)

	faasHandlers.Proxy = internalHandlers.MakeForwardingProxyHandler(reverseProxy, logger, metrics)
	faasHandlers.RoutelessProxy = internalHandlers.MakeForwardingProxyHandler(reverseProxy, logger, metrics)
	faasHandlers.ListFunctions = internalHandlers.MakeForwardingProxyHandler(reverseProxy, logger, metrics)
	faasHandlers.DeployFunction = internalHandlers.MakeForwardingProxyHandler(reverseProxy, logger, metrics)
	faasHandlers.DeleteFunction = internalHandlers.MakeForwardingProxyHandler(reverseProxy, logger, metrics)
	faasHandlers.UpdateFunction = internalHandlers.MakeForwardingProxyHandler(reverseProxy, logger, metrics)

	alertHandler := plugin.NewExternalServiceQuery(*config.FunctionsProviderURL)
	faasHandlers.Alert = internalHandlers.MakeAlertHandler(alertHandler)

	return faasHandlers
}

func setupSwarmProviderHandlers(config types.GatewayConfig, logger *logrus.Logger, metrics metrics.Metrics) types.HandlerSet {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		logger.Fatal("Error with Docker client.")
	}

	dockerVersion, err := dockerClient.ServerVersion(context.Background())
	if err != nil {
		logger.Fatal("Error with Docker server.\n", err)
	}

	logger.Printf("Docker API version: %s, %s\n", dockerVersion.APIVersion, dockerVersion.Version)

	var faasHandlers types.HandlerSet

	// How many times to reschedule a function.
	maxRestarts := uint64(5)

	// Delay between container restarts
	restartDelay := time.Second * 5

	faasHandlers.Proxy = internalHandlers.MakeProxy(metrics, true, dockerClient, logger)
	faasHandlers.RoutelessProxy = internalHandlers.MakeProxy(metrics, false, dockerClient, logger)
	faasHandlers.ListFunctions = internalHandlers.MakeFunctionReader(metrics, dockerClient)
	faasHandlers.DeployFunction = internalHandlers.MakeNewFunctionHandler(metrics, dockerClient, maxRestarts, restartDelay)
	faasHandlers.DeleteFunction = internalHandlers.MakeDeleteFunctionHandler(metrics, dockerClient)
	faasHandlers.UpdateFunction = internalHandlers.MakeUpdateFunctionHandler(metrics, dockerClient, maxRestarts, restartDelay)

	faasHandlers.Alert = internalHandlers.MakeAlertHandler(internalHandlers.NewSwarmServiceQuery(dockerClient))

	return faasHandlers
}

func setupReplicaWatcher(config types.GatewayConfig, logger *logrus.Logger, metrics metrics.Metrics) {
	servicePollInterval := time.Second * 5

	if config.UseExternalProvider() {
		replicas.AttachExternalWatcher(*config.FunctionsProviderURL, metrics, "func", servicePollInterval)
	} else {
		dockerClient, err := client.NewEnvClient()
		if err != nil {
			logger.Fatal("Error with Docker client.")
		}

		dockerVersion, err := dockerClient.ServerVersion(context.Background())
		if err != nil {
			logger.Fatal("Error with Docker server.\n", err)
		}

		logger.Printf("Docker API version: %s, %s\n", dockerVersion.APIVersion, dockerVersion.Version)
		functionLabel := "function"
		replicas.AttachSwarmWatcher(dockerClient, metrics, functionLabel, servicePollInterval)
	}
}

func setupNats(config types.GatewayConfig, logger *logrus.Logger, metrics metrics.Metrics, handlers *types.HandlerSet) {

	if config.UseNATS() {
		logger.Info("Async enabled: Using NATS Streaming.")
		natsQueue, queueErr := natsHandler.CreateNatsQueue(*config.NATSAddress, *config.NATSPort)
		if queueErr != nil {
			log.Fatalln(queueErr)
		}

		handlers.QueuedProxy = internalHandlers.MakeQueuedProxy(metrics, true, logger, natsQueue)
		handlers.AsyncReport = internalHandlers.MakeAsyncReport(logger, metrics)
	}
}

func setupRouter(handlers *types.HandlerSet) *mux.Router {
	r := mux.NewRouter()

	// r.StrictSlash(false)	// This didn't work, so register routes twice.
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", handlers.Proxy)
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", handlers.Proxy)

	r.HandleFunc("/system/alert", handlers.Alert)

	r.HandleFunc("/system/functions", handlers.ListFunctions).Methods("GET")
	r.HandleFunc("/system/functions", handlers.DeployFunction).Methods("POST")
	r.HandleFunc("/system/functions", handlers.DeleteFunction).Methods("DELETE")
	r.HandleFunc("/system/functions", handlers.UpdateFunction).Methods("PUT")

	if handlers.QueuedProxy != nil {
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}/", handlers.QueuedProxy).Methods("POST")
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}", handlers.QueuedProxy).Methods("POST")

		r.HandleFunc("/system/async-report", handlers.AsyncReport)
	}

	fs := http.FileServer(http.Dir("./assets/"))

	// This URL allows access from the UI to the OpenFaaS store
	allowedCORSHost := "raw.githubusercontent.com"
	fsCORS := internalHandlers.DecorateWithCORS(fs, allowedCORSHost)

	r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui", fsCORS)).Methods("GET")

	r.HandleFunc("/", handlers.RoutelessProxy).Methods("POST")

	metricsHandler := prometheus.Handler()
	r.Handle("/metrics", metricsHandler)

	r.Handle("/", http.RedirectHandler("/ui/", http.StatusMovedPermanently)).Methods("GET")

	return r
}

func startServer(config types.GatewayConfig, router *mux.Router) {
	tcpPort := 8080

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", tcpPort),
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes, // 1MB - can be overridden by setting Server.MaxHeaderBytes.
		Handler:        router,
	}

	log.Fatal(s.ListenAndServe())
}
