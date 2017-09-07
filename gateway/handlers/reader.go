// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

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

				// Ping counters
				// getCounterValue(service.Spec.Name, "200", &metricsOptions)
				// getCounterValue(service.Spec.Name, "500", &metricsOptions)

				var envProcess string

				for _, env := range service.Spec.TaskTemplate.ContainerSpec.Env {
					if strings.Index(env, "fprocess=") > -1 {
						envProcess = env[len("fprocess="):]
					}
				}

				f := requests.Function{
					Name:            service.Spec.Name,
					Image:           service.Spec.TaskTemplate.ContainerSpec.Image,
					InvocationCount: 0,
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

// func getCounterValue(service string, code string, metricsOptions *metrics.MetricOptions) float64 {

// 	metric, err := metricsOptions.GatewayFunctionInvocation.
// 		GetMetricWith(prometheus.Labels{"function_name": service, "code": code})

// 	if err != nil {
// 		return 0
// 	}

// 	// Get the metric's value from ProtoBuf interface (idea via Julius Volz)
// 	var protoMetric io_prometheus_client.Metric
// 	metric.Write(&protoMetric)
// 	invocations := protoMetric.GetCounter().GetValue()
// 	return invocations
// }
