package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/requests"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

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
