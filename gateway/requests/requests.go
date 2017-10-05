// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

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
}

// DeleteFunctionRequest delete a deployed function
type DeleteFunctionRequest struct {
	FunctionName string `json:"functionName"`
}

// PrometheusInnerAlertLabel PrometheusInnerAlertLabel
type PrometheusInnerAlertLabel struct {
	AlertName    string `json:"alertname"`
	FunctionName string `json:"function_name"`
}

// PrometheusInnerAlert PrometheusInnerAlert
type PrometheusInnerAlert struct {
	Status string                    `json:"status"`
	Labels PrometheusInnerAlertLabel `json:"labels"`
}

// PrometheusAlert as produced by AlertManager
type PrometheusAlert struct {
	Status   string                 `json:"status"`
	Receiver string                 `json:"receiver"`
	Alerts   []PrometheusInnerAlert `json:"alerts"`
}

// Function exported for system/functions endpoint
type Function struct {
	Name            string  `json:"name"`
	Image           string  `json:"image"`
	InvocationCount float64 `json:"invocationCount"` // TODO: shouldn't this be int64?
	Replicas        uint64  `json:"replicas"`
	EnvProcess      string  `json:"envProcess"`
}

// AsyncReport is the report from a function executed on a queue worker.
type AsyncReport struct {
	FunctionName string  `json:"name"`
	StatusCode   int     `json:"statusCode"`
	TimeTaken    float64 `json:"timeTaken"`
}
