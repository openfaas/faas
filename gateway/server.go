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
	"github.com/alexellis/faas/gateway/requests"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

func scaleService(req requests.PrometheusAlert, c *client.Client) error {
	var err error
	//Todo: convert to loop / handler.
	serviceName := req.Alerts[0].Labels.FunctionName
	service, _, inspectErr := c.ServiceInspectWithRaw(context.Background(), serviceName)
	if inspectErr != nil {
		var replicas uint64

		if req.Status == "firing" {
			if *service.Spec.Mode.Replicated.Replicas < 20 {
				replicas = *service.Spec.Mode.Replicated.Replicas + uint64(5)
			} else {
				return err
			}
		} else {
			replicas = *service.Spec.Mode.Replicated.Replicas - uint64(5)
			if replicas <= 0 {
				replicas = 1
			}
		}
		log.Printf("Scaling %s to %d replicas.\n", serviceName, replicas)

		service.Spec.Mode.Replicated.Replicas = &replicas
		updateOpts := types.ServiceUpdateOptions{}
		updateOpts.RegistryAuthFrom = types.RegistryAuthFromSpec

		response, updateErr := c.ServiceUpdate(context.Background(), service.ID, service.Version, service.Spec, updateOpts)
		if updateErr != nil {
			err = updateErr
		}
		log.Println(response)

	} else {
		err = inspectErr
	}

	return err
}

func makeAlertHandler(c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Alert received.")
		body, _ := ioutil.ReadAll(r.Body)

		var req requests.PrometheusAlert
		err := json.Unmarshal(body, &req)
		if err != nil {
			log.Println(err)
		}

		if len(req.Alerts) > 0 {
			err := scaleService(req, c)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}
	}
}

// makeFunctionReader gives a summary of Function structs with Docker service stats overlaid with Prometheus counters.
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
		functions := make([]requests.Function, 0)
		for _, service := range services {
			counter, _ := metricsOptions.GatewayFunctionInvocation.GetMetricWithLabelValues(service.Spec.Name)

			// Get the metric's value from ProtoBuf interface (idea via Julius Volz)
			var protoMetric io_prometheus_client.Metric
			counter.Write(&protoMetric)
			invocations := protoMetric.GetCounter().GetValue()

			f := requests.Function{
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
	dockerVersion, err := dockerClient.ServerVersion(context.Background())
	if err != nil {
		log.Fatal("Error with Docker server.\n", err)
	}
	log.Println("API version: %s, %s\n", dockerVersion.APIVersion, dockerVersion.Version)

	metricsOptions := metrics.BuildMetricsOptions()
	metrics.RegisterMetrics(metricsOptions)

	r := mux.NewRouter()
	r.HandleFunc("/function/{name:[a-zA-Z_]+}", MakeProxy(metricsOptions, true, dockerClient))
	r.HandleFunc("/system/alert", makeAlertHandler(dockerClient))
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
