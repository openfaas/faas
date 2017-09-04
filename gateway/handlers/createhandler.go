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

	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/requests"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/registry"
)

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
	linuxOnlyConstraints := []string{"node.platform.os == linux"}
	constraints := []string{}
	if request.Constraints != nil && len(request.Constraints) > 0 {
		constraints = request.Constraints
	} else {
		constraints = linuxOnlyConstraints
	}

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
			Placement: &swarm.Placement{
				Constraints: constraints,
			},
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
