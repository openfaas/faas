package types

import "time"

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

	// Usage represents CPU and RAM used by all of the
	// functions' replicas. Divide by AvailableReplicas for an
	// average value per replica.
	Usage *FunctionUsage `json:"usage,omitempty"`
}

// FunctionUsage represents CPU and RAM used by all of the
// functions' replicas.
//
// CPU is measured in seconds consumed since the last measurement
// RAM is measured in total bytes consumed
//
type FunctionUsage struct {
	// CPU is the increase in CPU usage since the last measurement
	// equivalent to Kubernetes' concept of millicores.
	CPU float64 `json:"cpu,omitempty"`

	//TotalMemoryBytes is the total memory usage in bytes.
	TotalMemoryBytes float64 `json:"totalMemoryBytes,omitempty"`
}
