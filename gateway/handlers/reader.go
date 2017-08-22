package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"strings"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/requests"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
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
		var functions []requests.Function

		for _, service := range services {

			if len(service.Spec.TaskTemplate.ContainerSpec.Labels["function"]) > 0 {
				invocations := getCounterValue(service.Spec.Name, "200", &metricsOptions) +
					getCounterValue(service.Spec.Name, "500", &metricsOptions)

				var envProcess string

				for _, env := range service.Spec.TaskTemplate.ContainerSpec.Env {
					if strings.Index(env, "fprocess=") > -1 {
						envProcess = env[len("fprocess="):]
					}
				}

				f := requests.Function{
					Name:            service.Spec.Name,
					Image:           service.Spec.TaskTemplate.ContainerSpec.Image,
					InvocationCount: invocations,
					Replicas:        *service.Spec.Mode.Replicated.Replicas,
					EnvProcess:      envProcess,
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
