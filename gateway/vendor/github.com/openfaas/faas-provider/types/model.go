package types

// FunctionDeployment represents a request to create or update a Function.
type FunctionDeployment struct {

	// Service corresponds to a Service
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

	// Namespace for the function to be deployed into
	Namespace string `json:"namespace,omitempty"`
}

// FunctionResources Memory and CPU
type FunctionResources struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

// FunctionStatus exported for system/functions endpoint
type FunctionStatus struct {

	// Name corresponds to a Service
	Name string `json:"name"`

	// Image corresponds to a Docker image
	Image string `json:"image"`

	// InvocationCount count of invocations
	InvocationCount float64 `json:"invocationCount"`

	// Replicas desired within the cluster
	Replicas uint64 `json:"replicas"`

	// EnvProcess is the process to pass to the watchdog, if in use
	EnvProcess string `json:"envProcess"`

	// AvailableReplicas is the count of replicas ready to receive
	// invocations as reported by the backend
	AvailableReplicas uint64 `json:"availableReplicas"`

	// Labels are metadata for functions which may be used by the
	// backend for making scheduling or routing decisions
	Labels *map[string]string `json:"labels"`

	// Annotations are metadata for functions which may be used by the
	// backend for management, orchestration, events and build tasks
	Annotations *map[string]string `json:"annotations"`

	// Namespace where the function can be accessed
	Namespace string `json:"namespace,omitempty"`
}

// Secret for underlying orchestrator
type Secret struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Value     string `json:"value,omitempty"`
}
