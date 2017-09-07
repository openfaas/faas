// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"fmt"

	"github.com/Sirupsen/logrus"
	natsHandler "github.com/alexellis/faas-nats/handler"
	internalHandlers "github.com/alexellis/faas/gateway/handlers"
	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/plugin"
	"github.com/alexellis/faas/gateway/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/gorilla/mux"
)

type handlerSet struct {
	Proxy          http.HandlerFunc
	DeployFunction http.HandlerFunc
	DeleteFunction http.HandlerFunc
	ListFunctions  http.HandlerFunc
	Alert          http.HandlerFunc
	RoutelessProxy http.HandlerFunc

	// QueuedProxy - queue work and return synchronous response
	QueuedProxy http.HandlerFunc

	// AsyncReport - report a defered execution result
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

		faasHandlers.Proxy = makeHandler(reverseProxy, &metricsOptions)
		faasHandlers.RoutelessProxy = makeHandler(reverseProxy, &metricsOptions)

		faasHandlers.Alert = internalHandlers.MakeAlertHandler(plugin.NewExternalServiceQuery(*config.FunctionsProviderURL))

		faasHandlers.ListFunctions = makeHandler(reverseProxy, &metricsOptions)
		faasHandlers.DeployFunction = makeHandler(reverseProxy, &metricsOptions)
		faasHandlers.DeleteFunction = makeHandler(reverseProxy, &metricsOptions)

		metrics.AttachExternalWatcher(*config.FunctionsProviderURL, metricsOptions, "func", time.Second*5)

	} else {
		maxRestarts := uint64(5)

		faasHandlers.Proxy = internalHandlers.MakeProxy(metricsOptions, true, dockerClient, &logger)
		faasHandlers.RoutelessProxy = internalHandlers.MakeProxy(metricsOptions, true, dockerClient, &logger)
		faasHandlers.Alert = internalHandlers.MakeAlertHandler(internalHandlers.NewSwarmServiceQuery(dockerClient))
		faasHandlers.ListFunctions = internalHandlers.MakeFunctionReader(metricsOptions, dockerClient)
		faasHandlers.DeployFunction = internalHandlers.MakeNewFunctionHandler(metricsOptions, dockerClient, maxRestarts)
		faasHandlers.DeleteFunction = internalHandlers.MakeDeleteFunctionHandler(metricsOptions, dockerClient)
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

	listFunctions := metrics.AddMetricsHandler(faasHandlers.ListFunctions, config.PrometheusHost, config.PrometheusPort)

	r := mux.NewRouter()

	// r.StrictSlash(false)	// This didn't work, so register routes twice.
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", faasHandlers.Proxy)
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", faasHandlers.Proxy)

	r.HandleFunc("/system/alert", faasHandlers.Alert)
	r.HandleFunc("/system/functions", listFunctions).Methods("GET")
	r.HandleFunc("/system/functions", faasHandlers.DeployFunction).Methods("POST")
	r.HandleFunc("/system/functions", faasHandlers.DeleteFunction).Methods("DELETE")

	if faasHandlers.QueuedProxy != nil {
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}/", faasHandlers.QueuedProxy).Methods("POST")
		r.HandleFunc("/async-function/{name:[-a-zA-Z_0-9]+}", faasHandlers.QueuedProxy).Methods("POST")

		r.HandleFunc("/system/async-report", faasHandlers.AsyncReport)
	}

	fs := http.FileServer(http.Dir("./assets/"))
	r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui", fs)).Methods("GET")

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

func makeHandler(proxy *httputil.ReverseProxy, metrics *metrics.MetricOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.String()

		log.Printf("Forwarding [%s] to %s", r.Method, r.URL.String())
		start := time.Now()

		writeAdapter := types.NewWriteAdapter(w)
		proxy.ServeHTTP(writeAdapter, r)

		seconds := time.Since(start).Seconds()
		fmt.Printf("[%d] took %f seconds\n", writeAdapter.GetHeaderCode(), seconds)

		forward := "/function/"
		if len(uri) > len(forward) && strings.Index(uri, forward) == 0 {
			fmt.Println("function=", uri[len(forward):])
			service := uri[len(forward):]
			metrics.GatewayFunctionsHistogram.WithLabelValues(service).Observe(seconds)
			code := writeAdapter.GetHeaderCode()
			metrics.GatewayFunctionInvocation.With(prometheus.Labels{"function_name": service, "code": strconv.Itoa(code)}).Inc()
		}
	}
}
