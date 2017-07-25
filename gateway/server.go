package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"

	"fmt"

	"github.com/Sirupsen/logrus"
	internalHandlers "github.com/alexellis/faas/gateway/handlers"
	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/types"
	"github.com/docker/docker/client"

	"github.com/gorilla/mux"
)

type handlerSet struct {
	Proxy          http.HandlerFunc
	DeployFunction http.HandlerFunc
	DeleteFunction http.HandlerFunc
	ListFunctions  http.HandlerFunc
	Alert          http.HandlerFunc
	RoutelessProxy http.HandlerFunc
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

		faasHandlers.Proxy = handler(reverseProxy)
		faasHandlers.RoutelessProxy = handler(reverseProxy)
		faasHandlers.Alert = handler(reverseProxy)
		faasHandlers.ListFunctions = handler(reverseProxy)
		faasHandlers.DeployFunction = handler(reverseProxy)
		faasHandlers.DeleteFunction = handler(reverseProxy)

	} else {
		faasHandlers.Proxy = internalHandlers.MakeProxy(metricsOptions, true, dockerClient, &logger)
		faasHandlers.RoutelessProxy = internalHandlers.MakeProxy(metricsOptions, true, dockerClient, &logger)
		faasHandlers.Alert = internalHandlers.MakeAlertHandler(dockerClient)
		faasHandlers.ListFunctions = internalHandlers.MakeFunctionReader(metricsOptions, dockerClient)
		faasHandlers.DeployFunction = internalHandlers.MakeNewFunctionHandler(metricsOptions, dockerClient)
		faasHandlers.DeleteFunction = internalHandlers.MakeDeleteFunctionHandler(metricsOptions, dockerClient)

		// This could exist in a separate process - records the replicas of each swarm service.
		functionLabel := "function"
		metrics.AttachSwarmWatcher(dockerClient, metricsOptions, functionLabel)
	}

	r := mux.NewRouter()

	// r.StrictSlash(false)	// This didn't work, so register routes twice.
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", faasHandlers.Proxy)
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", faasHandlers.Proxy)

	r.HandleFunc("/system/alert", faasHandlers.Alert)
	r.HandleFunc("/system/functions", faasHandlers.ListFunctions).Methods("GET")
	r.HandleFunc("/system/functions", faasHandlers.DeployFunction).Methods("POST")
	r.HandleFunc("/system/functions", faasHandlers.DeleteFunction).Methods("DELETE")

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

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Forwarding [%s] to %s", r.Method, r.URL.String())
		p.ServeHTTP(w, r)
	}
}
