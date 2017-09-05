// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/requests"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

// MakeFunctionReader gives a summary of Function structs with Docker service stats overlaid with Prometheus counters.
func MakeFunctionReader(metricsOptions metrics.MetricOptions, c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		client := http.Client{}

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

		req, reqErr := http.NewRequest("GET", "http://prometheus:9090/api/v1/query/?query=gateway_function_invocation_total", nil)
		if reqErr != nil {
			log.Println(reqErr)
		}
		res, getErr := client.Do(req)
		if getErr != nil {
			fmt.Fprintln(os.Stderr, getErr)
			return
		}
		defer res.Body.Close()
		bytesOut, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			fmt.Fprintln(os.Stderr, readErr)
		}

		fmt.Println(string(bytesOut))

		var values VectorQueryResponse

		unmarshalErr := json.Unmarshal(bytesOut, &values)
		if unmarshalErr != nil {
			fmt.Fprintln(os.Stderr, unmarshalErr)
		}

		for _, service := range services {

			if len(service.Spec.TaskTemplate.ContainerSpec.Labels["function"]) > 0 {

				var envProcess string

				for _, env := range service.Spec.TaskTemplate.ContainerSpec.Env {
					if strings.Index(env, "fprocess=") > -1 {
						envProcess = env[len("fprocess="):]
					}
				}

				var invocations float64
				for _, result := range values.Data.Result {
					if service.Spec.Name == result.Metric.FunctionName {
						val := result.Value[1]
						switch val.(type) {
						case string:
							f, strconvErr := strconv.ParseFloat(val.(string), 64)
							if err != nil {
								fmt.Fprintln(os.Stderr, strconvErr)
								continue
							}
							invocations = invocations + f
						}
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

func getCounterValue(service string, code string, metricsOptions *metrics.MetricOptions) float64 {

	metric, err := metricsOptions.GatewayFunctionInvocation.
		GetMetricWith(prometheus.Labels{"function_name": service, "code": code})

	if err != nil {
		return 0
	}

	// Get the metric's value from ProtoBuf interface (idea via Julius Volz)
	var protoMetric io_prometheus_client.Metric
	metric.Write(&protoMetric)
	invocations := protoMetric.GetCounter().GetValue()
	return invocations
}

type VectorQueryResponse struct {
	Data struct {
		Result []struct {
			Metric struct {
				Code         string `json:"code"`
				FunctionName string `json:"function_name"`
			}
			Value []interface{} `json:"value"`
		}
	}
}
