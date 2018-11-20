// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// Package requests package provides a client SDK or library for
// the OpenFaaS gateway REST API
package requests

// CreateFunctionRequest create a function in the swarm.
type CreateFunctionRequest struct {

	// Service corresponds to a Docker Service
	Service string `json:"service"`

	// Image corresponds to a Docker image
	Image string `json:"image"`

	// Network is specific to Docker Swarm - default overlay network is: func_functions
	Network string `json:"network"`

	// EnvProcess corresponds to the fprocess variable for your container watchdog.
	EnvProcess string `json:"envProcess"`

	// EnvVars provides overrides for functions.
	EnvVars map[string]string `json:"envVars"`

	// RegistryAuth is the registry authentication (optional)
	// in the same encoded format as Docker native credentials
	// (see ~/.docker/config.json)
	RegistryAuth string `json:"registryAuth,omitempty"`

	// Constraints are specific to back-end orchestration platform
	Constraints []string `json:"constraints"`

	// Secrets list of secrets to be made available to function
	Secrets []string `json:"secrets"`

	// Labels are metadata for functions which may be used by the
	// back-end for making scheduling or routing decisions
	Labels *map[string]string `json:"labels"`

	// Annotations are metadata for functions which may be used by the
	// back-end for management, orchestration, events and build tasks
	Annotations *map[string]string `json:"annotations"`

	// Limits for function
	Limits *FunctionResources `json:"limits"`

	// Requests of resources requested by function
	Requests *FunctionResources `json:"requests"`

	// ReadOnlyRootFilesystem removes write-access from the root filesystem
	// mount-point.
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem"`
}

// FunctionResources Memory and CPU
type FunctionResources struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

// Function exported for system/functions endpoint
type Function struct {
	Name            string  `json:"name"`
	Image           string  `json:"image"`
	InvocationCount float64 `json:"invocationCount"` // TODO: shouldn't this be int64?
	Replicas        uint64  `json:"replicas"`
	EnvProcess      string  `json:"envProcess"`

	// AvailableReplicas is the count of replicas ready to receive invocations as reported by the back-end
	AvailableReplicas uint64 `json:"availableReplicas"`

	// Labels are metadata for functions which may be used by the
	// back-end for making scheduling or routing decisions
	Labels *map[string]string `json:"labels"`

	// Annotations are metadata for functions which may be used by the
	// back-end for management, orchestration, events and build tasks
	Annotations *map[string]string `json:"annotations"`
}

// AsyncReport is the report from a function executed on a queue worker.
type AsyncReport struct {
	FunctionName string  `json:"name"`
	StatusCode   int     `json:"statusCode"`
	TimeTaken    float64 `json:"timeTaken"`
}

// DeleteFunctionRequest delete a deployed function
type DeleteFunctionRequest struct {
	FunctionName string `json:"functionName"`
}

// CreateSecretRequest create a secret w/ annotations
type CreateSecretRequest struct {
	Secret Secret `json:"secret"`
}

// DeleteSecretRequest remote a secret by name
type DeleteSecretRequest struct {
	SecretName string `json:"secretName"`
}

// Secret schema use Value only in write-only http verbs
type Secret struct {
	Name        string             `json:"name"`
	Value       string             `json:"value"` // write-only
	Annotations *map[string]string `json:"annotations"`
}
