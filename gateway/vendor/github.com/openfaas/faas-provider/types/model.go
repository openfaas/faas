package types

import "time"

// Secret for underlying orchestrator
type Secret struct {
	// Name of the secret
	Name string `json:"name"`

	// Namespace if applicable for the secret
	Namespace string `json:"namespace,omitempty"`

	// Value is a string representing the string's value
	Value string `json:"value,omitempty"`

	// RawValue can be used to provide binary data when
	// Value is not set
	RawValue []byte `json:"rawValue,omitempty"`
}

// FunctionDeployment represents a request to create or update a Function.
type FunctionDeployment struct {

	// Service is the name of the function deployment
	Service string `json:"service"`

	// Image is a fully-qualified container image
	Image string `json:"image"`

	// Namespace for the function, if supported by the faas-provider
	Namespace string `json:"namespace,omitempty"`

	// EnvProcess overrides the fprocess environment variable and can be used
	// with the watchdog
	EnvProcess string `json:"envProcess,omitempty"`

	// EnvVars can be provided to set environment variables for the function runtime.
	EnvVars map[string]string `json:"envVars,omitempty"`

	// Constraints are specific to the faas-provider.
	Constraints []string `json:"constraints,omitempty"`

	// Secrets list of secrets to be made available to function
	Secrets []string `json:"secrets,omitempty"`

	// Labels are metadata for functions which may be used by the
	// faas-provider or the gateway
	Labels *map[string]string `json:"labels,omitempty"`

	// Annotations are metadata for functions which may be used by the
	// faas-provider or the gateway
	Annotations *map[string]string `json:"annotations,omitempty"`

	// Limits for function
	Limits *FunctionResources `json:"limits,omitempty"`

	// Requests of resources requested by function
	Requests *FunctionResources `json:"requests,omitempty"`

	// ReadOnlyRootFilesystem removes write-access from the root filesystem
	// mount-point.
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem,omitempty"`
}

// FunctionResources Memory and CPU
type FunctionResources struct {
	Memory string `json:"memory,omitempty"`
	CPU    string `json:"cpu,omitempty"`
}

// FunctionStatus exported for system/functions endpoint
type FunctionStatus struct {

	// Name is the name of the function deployment
	Name string `json:"name"`

	// Image is a fully-qualified container image
	Image string `json:"image"`

	// Namespace for the function, if supported by the faas-provider
	Namespace string `json:"namespace,omitempty"`

	// EnvProcess overrides the fprocess environment variable and can be used
	// with the watchdog
	EnvProcess string `json:"envProcess,omitempty"`

	// EnvVars set environment variables for the function runtime
	EnvVars map[string]string `json:"envVars,omitempty"`

	// Constraints are specific to the faas-provider
	Constraints []string `json:"constraints,omitempty"`

	// Secrets list of secrets to be made available to function
	Secrets []string `json:"secrets,omitempty"`

	// Labels are metadata for functions which may be used by the
	// faas-provider or the gateway
	Labels *map[string]string `json:"labels,omitempty"`

	// Annotations are metadata for functions which may be used by the
	// faas-provider or the gateway
	Annotations *map[string]string `json:"annotations,omitempty"`

	// Limits for function
	Limits *FunctionResources `json:"limits,omitempty"`

	// Requests of resources requested by function
	Requests *FunctionResources `json:"requests,omitempty"`

	// ReadOnlyRootFilesystem removes write-access from the root filesystem
	// mount-point.
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem,omitempty"`

	// ================
	// Fields for status
	// ================

	// InvocationCount count of invocations
	InvocationCount float64 `json:"invocationCount,omitempty"`

	// Replicas desired within the cluster
	Replicas uint64 `json:"replicas,omitempty"`

	// AvailableReplicas is the count of replicas ready to receive
	// invocations as reported by the faas-provider
	AvailableReplicas uint64 `json:"availableReplicas,omitempty"`

	// CreatedAt is the time read back from the faas backend's
	// data store for when the function or its container was created.
	CreatedAt time.Time `json:"createdAt,omitempty"`
}
