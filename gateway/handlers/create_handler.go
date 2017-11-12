package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/registry"
	units "github.com/docker/go-units"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/requests"
)

var linuxOnlyConstraints = []string{"node.platform.os == linux"}

// MakeNewFunctionHandler creates a new function (service) inside the swarm network.
func MakeNewFunctionHandler(metricsOptions metrics.MetricOptions, c *client.Client, maxRestarts uint64, restartDelay time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			log.Println("Error parsing request:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		options := types.ServiceCreateOptions{}
		if len(request.RegistryAuth) > 0 {
			auth, err := BuildEncodedAuthConfig(request.RegistryAuth, request.Image)
			if err != nil {
				log.Println("Error building registry auth configuration:", err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid registry auth"))
				return
			}
			options.EncodedRegistryAuth = auth
		}

		spec := makeSpec(&request, maxRestarts, restartDelay)

		response, err := c.ServiceCreate(context.Background(), spec, options)
		if err != nil {
			log.Println("Error creating service:", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Deployment error: " + err.Error()))
			return
		}
		log.Println(response.ID, response.Warnings)
	}
}

func makeSpec(request *requests.CreateFunctionRequest, maxRestarts uint64, restartDelay time.Duration) swarm.ServiceSpec {
	constraints := []string{}

	if request.Constraints != nil && len(request.Constraints) > 0 {
		constraints = request.Constraints
	} else {
		constraints = linuxOnlyConstraints
	}

	labels := map[string]string{
		"com.openfaas.function": request.Service,
		"function":              "true", // backwards-compatible
	}

	if request.Labels != nil {
		for k, v := range *request.Labels {
			labels[k] = v
		}
	}

	resources := buildResources(request)

	nets := []swarm.NetworkAttachmentConfig{
		{
			Target: request.Network,
		},
	}

	spec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name:   request.Service,
			Labels: labels,
		},
		TaskTemplate: swarm.TaskSpec{
			RestartPolicy: &swarm.RestartPolicy{
				MaxAttempts: &maxRestarts,
				Condition:   swarm.RestartPolicyConditionAny,
				Delay:       &restartDelay,
			},
			ContainerSpec: swarm.ContainerSpec{
				Image:  request.Image,
				Labels: labels,
			},
			Networks:  nets,
			Resources: resources,
			Placement: &swarm.Placement{
				Constraints: constraints,
			},
		},
	}

	// TODO: request.EnvProcess should only be set if it's not nil, otherwise we override anything in the Docker image already
	env := buildEnv(request.EnvProcess, request.EnvVars)

	if len(env) > 0 {
		spec.TaskTemplate.ContainerSpec.Env = env
	}

	return spec
}

func buildEnv(envProcess string, envVars map[string]string) []string {
	var env []string
	if len(envProcess) > 0 {
		env = append(env, fmt.Sprintf("fprocess=%s", envProcess))
	}
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

// BuildEncodedAuthConfig for private registry
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

func ParseMemory(value string) (int64, error) {
	return units.RAMInBytes(value)
}

func buildResources(request *requests.CreateFunctionRequest) *swarm.ResourceRequirements {
	var resources *swarm.ResourceRequirements

	if request.Requests != nil || request.Limits != nil {

		resources = &swarm.ResourceRequirements{}
		if request.Limits != nil {
			memoryBytes, err := ParseMemory(request.Limits.Memory)
			if err != nil {
				log.Printf("Error parsing memory limit: %T", err)
			}

			resources.Limits = &swarm.Resources{
				MemoryBytes: memoryBytes,
			}
		}

		if request.Requests != nil {
			memoryBytes, err := ParseMemory(request.Requests.Memory)
			if err != nil {
				log.Printf("Error parsing memory request: %T", err)
			}

			resources.Reservations = &swarm.Resources{
				MemoryBytes: memoryBytes,
			}

		}
	}
	return resources
}
