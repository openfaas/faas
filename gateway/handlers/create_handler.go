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
	"strconv"
	"strings"
	"time"

	"github.com/docker/cli/opts"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
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

		secrets, err := makeSecretsArray(c, request.Secrets)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Deployment error: " + err.Error()))
			return
		}

		spec := makeSpec(&request, maxRestarts, restartDelay, secrets)

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

func makeSpec(request *requests.CreateFunctionRequest, maxRestarts uint64, restartDelay time.Duration, secrets []*swarm.SecretReference) swarm.ServiceSpec {
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
				Image:   request.Image,
				Labels:  labels,
				Secrets: secrets,
			},
			Networks:  nets,
			Resources: resources,
			Placement: &swarm.Placement{
				Constraints: constraints,
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: getMinReplicas(request),
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

func getMinReplicas(request *requests.CreateFunctionRequest) *uint64 {
	replicas := uint64(1)

	if request.Labels != nil {
		if val, exists := (*request.Labels)["com.openfaas.scale.min"]; exists {
			value, err := strconv.Atoi(val)
			if err != nil {
				log.Println(err)
			}
			replicas = uint64(value)
		}
	}
	return &replicas
}

func makeSecretsArray(c *client.Client, secretNames []string) ([]*swarm.SecretReference, error) {
	values := []*swarm.SecretReference{}

	if len(secretNames) == 0 {
		return values, nil
	}

	secretOpts := new(opts.SecretOpt)
	for _, secret := range secretNames {
		if err := secretOpts.Set(secret); err != nil {
			return nil, err
		}
	}

	requestedSecrets := make(map[string]bool)
	ctx := context.Background()

	// query the Swarm for the requested secret ids, these are required to complete
	// the spec
	args := filters.NewArgs()
	for _, opt := range secretOpts.Value() {
		args.Add("name", opt.SecretName)
	}

	secrets, err := c.SecretList(ctx, types.SecretListOptions{
		Filters: args,
	})
	if err != nil {
		return nil, err
	}

	// create map of matching secrets for easy lookup
	foundSecrets := make(map[string]string)
	for _, secret := range secrets {
		foundSecrets[secret.Spec.Annotations.Name] = secret.ID
	}

	// mimics the simple syntax for `docker service create --secret foo`
	// and the code is based on the docker cli
	for _, opts := range secretOpts.Value() {

		secretName := opts.SecretName
		if _, exists := requestedSecrets[secretName]; exists {
			return nil, fmt.Errorf("duplicate secret target for %s not allowed", secretName)
		}

		id, ok := foundSecrets[secretName]
		if !ok {
			return nil, fmt.Errorf("secret not found: %s; possible choices:\n%s", secretName, secrets)
		}

		options := new(swarm.SecretReference)
		*options = *opts
		options.SecretID = id

		requestedSecrets[secretName] = true
		values = append(values, options)
	}

	return values, nil
}
