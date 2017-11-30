// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
	internalHandlers "github.com/openfaas/faas/gateway/handlers"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/plugin"
	"github.com/openfaas/faas/gateway/types"
	natsHandler "github.com/openfaas/nats-queue-worker/handler"

	"github.com/gorilla/mux"
)

type handlerSet struct {
	Proxy          http.HandlerFunc
	DeployFunction http.HandlerFunc
	DeleteFunction http.HandlerFunc
	ListFunctions  http.HandlerFunc
	Alert          http.HandlerFunc
	RoutelessProxy http.HandlerFunc
	UpdateFunction http.HandlerFunc

	// QueuedProxy - queue work and return synchronous response
	QueuedProxy http.HandlerFunc

	// AsyncReport - report a deferred execution result
	AsyncReport http.HandlerFunc
}

func main() {
	logger := logrus.Logger{}
	logrus.SetFormatter(&logrus.TextFormatter{})

	osEnv := types.OsEnv{}
	readConfig := types.ReadConfig{}
	config := readConfig.Read(osEnv)

	log.Printf("HTTP Read Timeout: %s", config.ReadTimeout)
	log.Printf("HTTP Write Timeout: %s", config.WriteTimeout)

	var dockerClient *client.Client

	if config.UseExternalProvider() {
		log.Printf("Binding to external function provider: %s", config.FunctionsProviderURL)
	} else {
		var err error
		dockerClient, err = client.NewEnvClient()
		if err != nil {
			log.Fatal("Error with Docker client.")
		}
		dockerVersion, err := dockerClient.ServerVersion(context.Background())
		if err != nil {
			log.Fatal("Error with Docker server.\n", err)
		}
		log.Printf("Docker API version: %s, %s\n", dockerVersion.APIVersion, dockerVersion.Version)
	}

	metricsOptions := metrics.BuildMetricsOptions()
	metrics.RegisterMetrics(metricsOptions)

	var faasHandlers handlerSet

	if config.UseExternalProvider() {

		reverseProxy := httputil.NewSingleHostReverseProxy(config.FunctionsProviderURL)

		faasHandlers.Proxy = internalHandlers.MakeForwardingProxyHandler(reverseProxy, &metricsOptions)
		faasHandlers.RoutelessProxy = internalHandlers.MakeForwardingProxyHandler(reverseProxy, &metricsOptions)
		faasHandlers.ListFunctions = internalHandlers.MakeForwardingProxyHandler(reverseProxy, &metricsOptions)
		faasHandlers.DeployFunction = internalHandlers.MakeForwardingProxyHandler(reverseProxy, &metricsOptions)
		faasHandlers.DeleteFunction = internalHandlers.MakeForwardingProxyHandler(reverseProxy, &metricsOptions)
		faasHandlers.UpdateFunction = internalHandlers.MakeForwardingProxyHandler(reverseProxy, &metricsOptions)

		alertHandler := plugin.NewExternalServiceQuery(*config.FunctionsProviderURL)
		faasHandlers.Alert = internalHandlers.MakeAlertHandler(alertHandler)

		metrics.AttachExternalWatcher(*config.FunctionsProviderURL, metricsOptions, "func", time.Second*5)

	} else {

		// How many times to reschedule a function.
		maxRestarts := uint64(5)
		// Delay between container restarts
		restartDelay := time.Second * 5

		faasHandlers.Proxy = internalHandlers.MakeProxy(metricsOptions, true, dockerClient, &logger)
		faasHandlers.RoutelessProxy = internalHandlers.MakeProxy(metricsOptions, false, dockerClient, &logger)
		faasHandlers.ListFunctions = internalHandlers.MakeFunctionReader(metricsOptions, dockerClient)
		faasHandlers.DeployFunction = internalHandlers.MakeNewFunctionHandler(metricsOptions, dockerClient, maxRestarts, restartDelay)
		faasHandlers.DeleteFunction = internalHandlers.MakeDeleteFunctionHandler(metricsOptions, dockerClient)
		faasHandlers.UpdateFunction = internalHandlers.MakeUpdateFunctionHandler(metricsOptions, dockerClient, maxRestarts, restartDelay)

		faasHandlers.Alert = internalHandlers.MakeAlertHandler(internalHandlers.NewSwarmServiceQuery(dockerClient))

		// This could exist in a separate process - records the replicas of each swarm service.
		functionLabel := "function"
		metrics.AttachSwarmWatcher(dockerClient, metricsOptions, functionLabel)
	}

	if config.UseNATS() {
		log.Println("Async enabled: Using NATS Streaming.")
		natsQueue, queueErr := natsHandler.CreateNatsQueue(*config.NATSAddress, *config.NATSPort)
		if queueErr != nil {
			log.Fatalln(queueErr)
		}

		faasHandlers.QueuedProxy = internalHandlers.MakeQueuedProxy(metricsOptions, true, &logger, natsQueue)
		faasHandlers.AsyncReport = internalHandlers.MakeAsyncReport(metricsOptions)
	}

	prometheusQuery := metrics.NewPrometheusQuery(config.PrometheusHost, config.PrometheusPort, &http.Client{})
	listFunctions := metrics.AddMetricsHandler(faasHandlers.ListFunctions, prometheusQuery)

	r := mux.NewRouter()

	// r.StrictSlash(false)	// This didn't work, so register routes twice.
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", faasHandlers.Proxy)
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", faasHandlers.Proxy)

	r.HandleFunc("/system/alert", faasHandlers.Alert)
	r.HandleFunc("/system/functions", listFunctions).Methods("GET")
	r.HandleFunc("/system/functions", faasHandlers.DeployFunction).Methods("POST")
	r.HandleFunc("/system/functions", faasHandlers.DeleteFunction).Methods("DELETE")
	r.HandleFunc("/system/functions", faasHandlers.UpdateFunction).Methods("PUT")

	if faasHandlers.QueuedProxy != nil {
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}/", faasHandlers.QueuedProxy).Methods("POST")
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}", faasHandlers.QueuedProxy).Methods("POST")

		r.HandleFunc("/system/async-report", faasHandlers.AsyncReport)
	}

	fs := http.FileServer(http.Dir("./assets/"))

	// This URL allows access from the UI to the OpenFaaS store
	allowedCORSHost := "raw.githubusercontent.com"
	fsCORS := internalHandlers.DecorateWithCORS(fs, allowedCORSHost)

	r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui", fsCORS)).Methods("GET")

	r.HandleFunc("/", faasHandlers.RoutelessProxy).Methods("POST")

	metricsHandler := metrics.PrometheusHandler()
	r.Handle("/metrics", metricsHandler)
	r.Handle("/", http.RedirectHandler("/ui/", http.StatusMovedPermanently)).Methods("GET")

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
