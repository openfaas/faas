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
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/requests"
)

// MakeFunctionReader gives a summary of Function structs with Docker service stats overlaid with Prometheus counters.
func MakeFunctionReader(metricsOptions metrics.MetricOptions, serviceClient client.ServiceAPIClient, nodeClient client.NodeAPIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		verbose := false

		if r.URL != nil {
			verbose = queryIsNotFalse(r, "v")
		}
		serviceFilter := filters.NewArgs()

		options := types.ServiceListOptions{
			Filters: serviceFilter,
		}

		ctx := context.Background()
		services, err := serviceClient.ServiceList(ctx, options)
		if err != nil {
			log.Println("Error getting service list:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error getting service list"))
			return
		}

		// TODO: Filter only "faas" functions (via metadata?)
		functions := []requests.Function{}

		var running map[string]int
		if verbose {
			running = getReplicaInfo(serviceClient, nodeClient, services, ctx)
		}

		for _, service := range services {

			if len(service.Spec.TaskTemplate.ContainerSpec.Labels["function"]) > 0 {
				envProcess := getEnvProcess(service.Spec.TaskTemplate.ContainerSpec.Env)

				// Required (copy by value)
				labels := service.Spec.Annotations.Labels

				var availableReplicas int
				if _, ok := running[service.ID]; ok {
					availableReplicas = running[service.ID]
				}

				f := requests.Function{
					Name:              service.Spec.Name,
					Image:             service.Spec.TaskTemplate.ContainerSpec.Image,
					InvocationCount:   0,
					Replicas:          *service.Spec.Mode.Replicated.Replicas,
					AvailableReplicas: availableReplicas,
					EnvProcess:        envProcess,
					Labels:            &labels,
				}

				functions = append(functions, f)
			}
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)

	}
}

func getEnvProcess(envVars []string) string {
	var value string
	for _, env := range envVars {
		if strings.Contains(env, "fprocess=") {
			value = env[len("fprocess="):]
		}
	}

	return value
}

func getReplicaInfo(serviceClient client.ServiceAPIClient, nodeClient client.NodeAPIClient, services []swarm.Service, ctx context.Context) map[string]int {
	// Begin replica info section
	taskFilter := filters.NewArgs()

	taskFilter.Add("_up-to-date", "true")
	tasks, err := serviceClient.TaskList(ctx, types.TaskListOptions{Filters: taskFilter})
	if err != nil {
		log.Println(err)
	}

	nodes, err := nodeClient.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		log.Println(err)
	}

	activeNodes := make(map[string]struct{})
	for _, n := range nodes {
		if n.Status.State != swarm.NodeStateDown {
			activeNodes[n.ID] = struct{}{}
		}
	}

	running := map[string]int{}
	for _, task := range tasks {
		if _, nodeActive := activeNodes[task.NodeID]; nodeActive && task.Status.State == swarm.TaskStateRunning {
			running[task.ServiceID]++
		}
	}

	//log.Printf("running: %v\n", running)
	return running
}

func queryIsNotFalse(r *http.Request, k string) bool {
	s := strings.ToLower(strings.TrimSpace(r.FormValue(k)))
	return !(s == "" || s == "0" || s == "no" || s == "false" || s == "none")
}
