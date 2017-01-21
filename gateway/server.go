package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
)

func makeAlertHandler(c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(c)
		// Todo: parse alert, validate alert and scale up or down function

		fmt.Println("Alert received.")
	}
}

func main() {
	var dockerClient *client.Client
	var err error
	dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Fatal("Error with Docker client.")
	}

	metricsOptions := metrics.BuildMetricsOptions()
	metrics.RegisterMetrics(metricsOptions)

	r := mux.NewRouter()
	r.HandleFunc("/function/{name:[a-zA-Z_]+}", MakeProxy(metricsOptions, true, dockerClient))
	r.HandleFunc("/system/alert", makeAlertHandler(dockerClient))
	r.HandleFunc("/", MakeProxy(metricsOptions, false, dockerClient))

	metricsHandler := metrics.PrometheusHandler()
	r.Handle("/metrics", metricsHandler)

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    8 * time.Second,
		WriteTimeout:   8 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        r,
	}

	log.Fatal(s.ListenAndServe())
}
