package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"io/ioutil"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/requests"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

func getCounterValue(service string, code string, metricsOptions *metrics.MetricOptions) float64 {

	metric, err := metricsOptions.GatewayFunctionInvocation.GetMetricWith(prometheus.Labels{"function_name": service, "code": code})
	if err != nil {
		return 0
	}

	// Get the metric's value from ProtoBuf interface (idea via Julius Volz)
	var protoMetric io_prometheus_client.Metric
	metric.Write(&protoMetric)
	invocations := protoMetric.GetCounter().GetValue()
	return invocations
}

// MakeFunctionReader gives a summary of Function structs with Docker service stats overlaid with Prometheus counters.
func MakeFunctionReader(metricsOptions metrics.MetricOptions, c *client.Client) http.HandlerFunc {
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
		var functions []requests.Function

		for _, service := range services {

			if len(service.Spec.TaskTemplate.ContainerSpec.Labels["function"]) > 0 {
				invocations := getCounterValue(service.Spec.Name, "200", &metricsOptions) +
					getCounterValue(service.Spec.Name, "500", &metricsOptions)

				f := requests.Function{
					Name:            service.Spec.Name,
					Image:           service.Spec.TaskTemplate.ContainerSpec.Image,
					InvocationCount: invocations,
					Replicas:        *service.Spec.Mode.Replicated.Replicas,
				}
				functions = append(functions, f)
			}
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)
	}
}

// MakeNewFunctionHandler creates a new function (service) inside the swarm network.
func MakeNewFunctionHandler(metricsOptions metrics.MetricOptions, c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Println(request)
		w.WriteHeader(http.StatusNotImplemented)
		options := types.ServiceCreateOptions{}
		spec := makeSpec(&request)

		response, err := c.ServiceCreate(context.Background(), spec, options)
		if err != nil {
			log.Println(err)
		}
		log.Println(response.ID, response.Warnings)
	}
}

func makeSpec(request *requests.CreateFunctionRequest) swarm.ServiceSpec {
	max := uint64(1)

	nets := []swarm.NetworkAttachmentConfig{
		swarm.NetworkAttachmentConfig{Target: request.Network},
	}

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			RestartPolicy: &swarm.RestartPolicy{
				MaxAttempts: &max,
				Condition:   swarm.RestartPolicyConditionNone,
			},
			ContainerSpec: swarm.ContainerSpec{
				Image:  request.Image,
				Env:    []string{fmt.Sprintf("fprocess=%s", request.EnvProcess)},
				Labels: map[string]string{"function": "true"},
			},
			Networks: nets,
		},
		Annotations: swarm.Annotations{
			Name: request.Service,
		},
	}
	return spec
}
