// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"io/ioutil"

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/requests"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/registry"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

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

func MakeDeleteFunctionHandler(metricsOptions metrics.MetricOptions, c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		req := requests.DeleteFunctionRequest{}
		defer r.Body.Close()
		reqData, _ := ioutil.ReadAll(r.Body)
		unmarshalErr := json.Unmarshal(reqData, &req)

		if (len(req.FunctionName) == 0) || unmarshalErr != nil {
			log.Printf("Error parsing request to remove service: %s\n", unmarshalErr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Attempting to remove service %s\n", req.FunctionName)

		serviceFilter := filters.NewArgs()
		options := types.ServiceListOptions{
			Filters: serviceFilter,
		}

		services, err := c.ServiceList(context.Background(), options)
		if err != nil {
			fmt.Println(err)
		}

		// TODO: Filter only "faas" functions (via metadata?)
		var serviceIDs []string
		for _, service := range services {
			isFunction := len(service.Spec.TaskTemplate.ContainerSpec.Labels["function"]) > 0

			if isFunction && req.FunctionName == service.Spec.Name {
				serviceIDs = append(serviceIDs, service.ID)
			}
		}

		log.Println(len(serviceIDs))
		if len(serviceIDs) == 0 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("No such service found: %s.", req.FunctionName)))
			return
		}

		var serviceRemoveErrors []error
		for _, serviceID := range serviceIDs {
			err := c.ServiceRemove(context.Background(), serviceID)
			if err != nil {
				serviceRemoveErrors = append(serviceRemoveErrors, err)
			}
		}

		if len(serviceRemoveErrors) > 0 {
			log.Printf("Error(s) removing service: %s\n", req.FunctionName)
			log.Println(serviceRemoveErrors)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}

	}
}

// MakeNewFunctionHandler creates a new function (service) inside the swarm network.
func MakeNewFunctionHandler(metricsOptions metrics.MetricOptions, c *client.Client, maxRestarts uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Println(request)

		// TODO: review why this was here... debugging?
		// w.WriteHeader(http.StatusNotImplemented)

		options := types.ServiceCreateOptions{}
		if len(request.RegistryAuth) > 0 {
			auth, err := BuildEncodedAuthConfig(request.RegistryAuth, request.Image)
			if err != nil {
				log.Println("Error while building registry auth configuration", err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid registry auth"))
				return
			}
			options.EncodedRegistryAuth = auth
		}
		spec := makeSpec(&request, maxRestarts)

		response, err := c.ServiceCreate(context.Background(), spec, options)
		if err != nil {
			log.Println(err)
		}
		log.Println(response.ID, response.Warnings)
	}
}

func makeSpec(request *requests.CreateFunctionRequest, maxRestarts uint64) swarm.ServiceSpec {

	nets := []swarm.NetworkAttachmentConfig{
		{Target: request.Network},
	}
	restartDelay := time.Second * 5

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			RestartPolicy: &swarm.RestartPolicy{
				MaxAttempts: &maxRestarts,
				Condition:   swarm.RestartPolicyConditionAny,
				Delay:       &restartDelay,
			},
			ContainerSpec: swarm.ContainerSpec{
				Image:  request.Image,
				Labels: map[string]string{"function": "true"},
			},
			Networks: nets,
		},
		Annotations: swarm.Annotations{
			Name: request.Service,
		},
	}

	// TODO: request.EnvProcess should only be set if it's not nil, otherwise we override anything in the Docker image already
	var env []string
	if len(request.EnvProcess) > 0 {
		env = append(env, fmt.Sprintf("fprocess=%s", request.EnvProcess))
	}
	for k, v := range request.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	if len(env) > 0 {
		spec.TaskTemplate.ContainerSpec.Env = env
	}

	return spec
}

func BuildEncodedAuthConfig(basicAuthB64 string, dockerImage string) (string, error) {
	// extract registry server address
	distributionRef, err := reference.ParseNormalizedNamed(dockerImage)
	if err != nil {
		return "", err
	}
	repoInfo, err := registry.ParseRepositoryInfo(distributionRef)
	if err != nil {
		return "", err
	}
	// extract registry user & password
	user, password, err := userPasswordFromBasicAuth(basicAuthB64)
	if err != nil {
		return "", err
	}
	// build encoded registry auth config
	buf, err := json.Marshal(types.AuthConfig{
		Username:      user,
		Password:      password,
		ServerAddress: repoInfo.Index.Name,
	})
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}

func userPasswordFromBasicAuth(basicAuthB64 string) (string, string, error) {
	c, err := base64.StdEncoding.DecodeString(basicAuthB64)
	if err != nil {
		return "", "", err
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return "", "", errors.New("Invalid basic auth")
	}
	return cs[:s], cs[s+1:], nil
}
