package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	faashandlers "github.com/alexellis/faas/gateway/handlers"
	"github.com/alexellis/faas/gateway/metrics"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
)

func main() {
	logger := logrus.Logger{}
	logrus.SetFormatter(&logrus.TextFormatter{})

	var dockerClient *client.Client
	var err error
	dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Fatal("Error with Docker client.")
	}
	dockerVersion, err := dockerClient.ServerVersion(context.Background())
	if err != nil {
		log.Fatal("Error with Docker server.\n", err)
	}
	log.Printf("API version: %s, %s\n", dockerVersion.APIVersion, dockerVersion.Version)

	metricsOptions := metrics.BuildMetricsOptions()
	metrics.RegisterMetrics(metricsOptions)

	r := mux.NewRouter()
	// r.StrictSlash(false)

	functionHandler := faashandlers.MakeProxy(metricsOptions, true, dockerClient, &logger)
	r.HandleFunc("/function/{name:[a-zA-Z_]+}", functionHandler)
	r.HandleFunc("/function/{name:[a-zA-Z_]+}/", functionHandler)

	r.HandleFunc("/system/alert", faashandlers.MakeAlertHandler(dockerClient))
	r.HandleFunc("/system/functions", faashandlers.MakeFunctionReader(metricsOptions, dockerClient)).Methods("GET")

	r.HandleFunc("/", faashandlers.MakeProxy(metricsOptions, false, dockerClient, &logger)).Methods("POST")

	metricsHandler := metrics.PrometheusHandler()
	r.Handle("/metrics", metricsHandler)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./assets/"))).Methods("GET")
	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    8 * time.Second,
		WriteTimeout:   8 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        r,
	}

	log.Fatal(s.ListenAndServe())
}
