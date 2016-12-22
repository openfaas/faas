package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"time"

	"io/ioutil"

	"strconv"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

func lookupSwarmService(serviceName string) (bool, error) {
	var c *client.Client
	var err error
	c, err = client.NewEnvClient()
	if err != nil {
		log.Fatal("Error with Docker client.")
	}
	serviceFilter := filters.NewArgs()
	serviceFilter.Add("name", serviceName)
	services, err := c.ServiceList(context.Background(), types.ServiceListOptions{Filters: serviceFilter})

	return len(services) > 0, err
}

func makeProxy(metrics metrics.MetricOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.GatewayRequestsTotal.Inc()

		start := time.Now()

		if r.Method == "POST" {
			log.Println(r.Header)
			header := r.Header["X-Function"]
			log.Println(header)

			exists, err := lookupSwarmService(header[0])
			if err != nil {
				log.Fatalln(err)
			}

			if exists == true {
				requestBody, _ := ioutil.ReadAll(r.Body)
				buf := bytes.NewBuffer(requestBody)

				response, err := http.Post("http://"+header[0]+":"+strconv.Itoa(8080)+"/", "text/plain", buf)
				if err != nil {
					log.Fatalln(err)
				}
				responseBody, _ := ioutil.ReadAll(response.Body)
				w.Write(responseBody)
				metrics.GatewayServerlessServedTotal.Inc()

				metrics.GatewayFunctions.Observe(time.Since(start).Seconds())
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Provide an x-function header."))
			}
		}
	}
}

func main() {
	GatewayRequestsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_requests_total",
		Help: "Total amount of HTTP requests to the gateway",
	})
	GatewayServerlessServedTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_serverless_invocation_total",
		Help: "Total amount of serverless function invocations",
	})
	GatewayFunctions := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "gateway_functions",
		Help: "Gateway functions",
	})

	prometheus.Register(GatewayRequestsTotal)
	prometheus.Register(GatewayServerlessServedTotal)
	prometheus.Register(GatewayFunctions)

	r := mux.NewRouter()
	r.HandleFunc("/", makeProxy(metrics.MetricOptions{
		GatewayRequestsTotal:         GatewayRequestsTotal,
		GatewayServerlessServedTotal: GatewayServerlessServedTotal,
		GatewayFunctions:             GatewayFunctions,
	}))

	metricsHandler := metrics.PrometheusHandler()
	r.Handle("/metrics", metricsHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
