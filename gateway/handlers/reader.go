// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/requests"
)

// MakeFunctionReader gives a summary of Function structs with Docker service stats overlaid with Prometheus counters.
func MakeFunctionReader(metricsOptions metrics.MetricOptions, c client.ServiceAPIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		serviceFilter := filters.NewArgs()

		options := types.ServiceListOptions{
			Filters: serviceFilter,
		}

		services, err := c.ServiceList(context.Background(), options)
		if err != nil {
			log.Println("Error getting service list:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error getting service list"))
			return
		}

		// TODO: Filter only "faas" functions (via metadata?)
		functions := []requests.Function{}

		for _, service := range services {

			if len(service.Spec.TaskTemplate.ContainerSpec.Labels["function"]) > 0 {
				var envProcess string

				for _, env := range service.Spec.TaskTemplate.ContainerSpec.Env {
					if strings.Contains(env, "fprocess=") {
						envProcess = env[len("fprocess="):]
					}
				}

				// Required (copy by value)
				labels := service.Spec.Annotations.Labels
				f := requests.Function{
					Name:            service.Spec.Name,
					Image:           service.Spec.TaskTemplate.ContainerSpec.Image,
					InvocationCount: 0,
					Replicas:        *service.Spec.Mode.Replicated.Replicas,
					EnvProcess:      envProcess,
					Labels:          &labels,
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
