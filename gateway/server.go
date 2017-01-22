package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

func makeAlertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Alert received.")
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(body))
		// Todo: parse alert, validate alert and scale up or down function

		w.WriteHeader(http.StatusOK)
	}
}

// Function exported for system/functions endpoint
type Function struct {
	Name            string  `json:"name"`
	Image           string  `json:"image"`
	InvocationCount float64 `json:"invocationCount"`
	Replicas        uint64  `json:"replicas"`
}

func makeFunctionReader(metricsOptions metrics.MetricOptions, c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		serviceFilter := filters.NewArgs()

		options := types.ServiceListOptions{
			Filters: serviceFilter,
		}

		services, err := c.ServiceList(context.Background(), options)
		if err != nil {
			fmt.Println(err)
		}

		// TODO: Filter only "faas" functions (via metadata?)

		functions := make([]Function, 0)
		for _, service := range services {
			counter, _ := metricsOptions.GatewayFunctionInvocation.GetMetricWithLabelValues(service.Spec.Name)

			var pbmetric io_prometheus_client.Metric
			counter.Write(&pbmetric)
			invocations := pbmetric.GetCounter().GetValue()

			f := Function{
				Name:            service.Spec.Name,
				Image:           service.Spec.TaskTemplate.ContainerSpec.Image,
				InvocationCount: invocations,
				Replicas:        *service.Spec.Mode.Replicated.Replicas,
			}
			functions = append(functions, f)
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)
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
	r.HandleFunc("/system/alert", makeAlertHandler())
	r.HandleFunc("/system/functions", makeFunctionReader(metricsOptions, dockerClient)).Methods("GET")
	r.HandleFunc("/", MakeProxy(metricsOptions, false, dockerClient)).Methods("POST")

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
