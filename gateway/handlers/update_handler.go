package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/requests"
)

// MakeUpdateFunctionHandler request to update an existing function with new configuration such as image, envvars etc.
func MakeUpdateFunctionHandler(metricsOptions metrics.MetricOptions, c *client.Client, maxRestarts uint64, restartDelay time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			log.Println("Error parsing request:", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		serviceInspectopts := types.ServiceInspectOptions{
			InsertDefaults: true,
		}

		service, _, err := c.ServiceInspectWithRaw(ctx, request.Service, serviceInspectopts)
		if err != nil {
			log.Println("Error inspecting service", err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}

		updateSpec(&request, &service.Spec, maxRestarts, restartDelay)

		updateOpts := types.ServiceUpdateOptions{}
		updateOpts.RegistryAuthFrom = types.RegistryAuthFromSpec

		if len(request.RegistryAuth) > 0 {
			auth, err := BuildEncodedAuthConfig(request.RegistryAuth, request.Image)
			if err != nil {
				log.Println("Error building registry auth configuration:", err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid registry auth"))
				return
			}
			updateOpts.EncodedRegistryAuth = auth
		}

		response, err := c.ServiceUpdate(ctx, service.ID, service.Version, service.Spec, updateOpts)
		if err != nil {
			log.Println("Error updating service:", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Update error: " + err.Error()))
			return
		}
		log.Println(response.Warnings)
	}
}

func updateSpec(request *requests.CreateFunctionRequest, spec *swarm.ServiceSpec, maxRestarts uint64, restartDelay time.Duration) {

	constraints := []string{}
	if request.Constraints != nil && len(request.Constraints) > 0 {
		constraints = request.Constraints
	} else {
		constraints = linuxOnlyConstraints
	}

	spec.TaskTemplate.RestartPolicy.MaxAttempts = &maxRestarts
	spec.TaskTemplate.RestartPolicy.Condition = swarm.RestartPolicyConditionAny
	spec.TaskTemplate.RestartPolicy.Delay = &restartDelay
	spec.TaskTemplate.ContainerSpec.Image = request.Image
	spec.TaskTemplate.ContainerSpec.Labels = map[string]string{
		"function":              "true",
		"com.openfaas.function": request.Service,
		"com.openfaas.uid":      fmt.Sprintf("%d", time.Now().Nanosecond()),
	}

	if request.Labels != nil {
		for k, v := range *request.Labels {
			spec.TaskTemplate.ContainerSpec.Labels[k] = v
			spec.Annotations.Labels[k] = v
		}
	}

	spec.TaskTemplate.Networks = []swarm.NetworkAttachmentConfig{
		{
			Target: request.Network,
		},
	}

	spec.TaskTemplate.Resources = buildResources(request)

	spec.TaskTemplate.Placement = &swarm.Placement{
		Constraints: constraints,
	}

	spec.Annotations.Name = request.Service

	spec.RollbackConfig = &swarm.UpdateConfig{
		FailureAction: "pause",
	}

	spec.UpdateConfig = &swarm.UpdateConfig{
		Parallelism:   1,
		FailureAction: "rollback",
	}

	env := buildEnv(request.EnvProcess, request.EnvVars)

	if len(env) > 0 {
		spec.TaskTemplate.ContainerSpec.Env = env
	}
}
