// This file is auto-generated, DO NOT EDIT.
//
// Source:
//
//	Title: OpenFaaS API Gateway
//	Version: 0.8.12
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// FunctionDeployment is an object.
type FunctionDeployment struct {
	// Annotations: A map of annotations for management, orchestration, events and build tasks
	Annotations map[string]string `json:"annotations,omitempty" mapstructure:"annotations,omitempty"`
	// Constraints:
	Constraints []string `json:"constraints,omitempty" mapstructure:"constraints,omitempty"`
	// EnvProcess: Process for watchdog to fork, i.e. the command to start the function process.
	//
	// This value configures the `fprocess` env variable.
	EnvProcess string `json:"envProcess,omitempty" mapstructure:"envProcess,omitempty"`
	// EnvVars: Overrides to environmental variables
	EnvVars map[string]string `json:"envVars,omitempty" mapstructure:"envVars,omitempty"`
	// Image: Docker image in accessible registry
	Image string `json:"image" mapstructure:"image"`
	// Labels: A map of labels for making scheduling or routing decisions
	Labels map[string]string `json:"labels,omitempty" mapstructure:"labels,omitempty"`
	// Limits:
	Limits *FunctionResources `json:"limits,omitempty" mapstructure:"limits,omitempty"`
	// Namespace: Namespace to deploy function to. When omitted, the default namespace is used, typically this is `openfaas-fn` but is configured by the provider.
	Namespace string `json:"namespace,omitempty" mapstructure:"namespace,omitempty"`
	// Network: Deprecated: Network, usually func_functions for Swarm.
	//
	// This value is completely ignored.
	Network string `json:"network,omitempty" mapstructure:"network,omitempty"`
	// ReadOnlyRootFilesystem: Make the root filesystem of the function read-only
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem,omitempty" mapstructure:"readOnlyRootFilesystem,omitempty"`
	// RegistryAuth: Deprecated: Private registry base64-encoded basic auth (as present in ~/.docker/config.json)
	//
	// Use a Kubernetes Secret with registry-auth secret type to provide this value instead.
	//
	// This value is completely ignored.
	RegistryAuth string `json:"registryAuth,omitempty" mapstructure:"registryAuth,omitempty"`
	// Requests:
	Requests *FunctionResources `json:"requests,omitempty" mapstructure:"requests,omitempty"`
	// Secrets:
	Secrets []string `json:"secrets,omitempty" mapstructure:"secrets,omitempty"`
	// Service: Name of deployed function
	Service string `json:"service" mapstructure:"service"`
}

// Validate implements basic validation for this model
func (m FunctionDeployment) Validate() error {
	return validation.Errors{
		"annotations": validation.Validate(
			m.Annotations,
		),
		"constraints": validation.Validate(
			m.Constraints,
		),
		"envVars": validation.Validate(
			m.EnvVars,
		),
		"labels": validation.Validate(
			m.Labels,
		),
		"limits": validation.Validate(
			m.Limits,
		),
		"requests": validation.Validate(
			m.Requests,
		),
		"secrets": validation.Validate(
			m.Secrets,
		),
	}.Filter()
}

// GetAnnotations returns the Annotations property
func (m FunctionDeployment) GetAnnotations() map[string]string {
	return m.Annotations
}

// SetAnnotations sets the Annotations property
func (m *FunctionDeployment) SetAnnotations(val map[string]string) {
	m.Annotations = val
}

// GetConstraints returns the Constraints property
func (m FunctionDeployment) GetConstraints() []string {
	return m.Constraints
}

// SetConstraints sets the Constraints property
func (m *FunctionDeployment) SetConstraints(val []string) {
	m.Constraints = val
}

// GetEnvProcess returns the EnvProcess property
func (m FunctionDeployment) GetEnvProcess() string {
	return m.EnvProcess
}

// SetEnvProcess sets the EnvProcess property
func (m *FunctionDeployment) SetEnvProcess(val string) {
	m.EnvProcess = val
}

// GetEnvVars returns the EnvVars property
func (m FunctionDeployment) GetEnvVars() map[string]string {
	return m.EnvVars
}

// SetEnvVars sets the EnvVars property
func (m *FunctionDeployment) SetEnvVars(val map[string]string) {
	m.EnvVars = val
}

// GetImage returns the Image property
func (m FunctionDeployment) GetImage() string {
	return m.Image
}

// SetImage sets the Image property
func (m *FunctionDeployment) SetImage(val string) {
	m.Image = val
}

// GetLabels returns the Labels property
func (m FunctionDeployment) GetLabels() map[string]string {
	return m.Labels
}

// SetLabels sets the Labels property
func (m *FunctionDeployment) SetLabels(val map[string]string) {
	m.Labels = val
}

// GetLimits returns the Limits property
func (m FunctionDeployment) GetLimits() *FunctionResources {
	return m.Limits
}

// SetLimits sets the Limits property
func (m *FunctionDeployment) SetLimits(val *FunctionResources) {
	m.Limits = val
}

// GetNamespace returns the Namespace property
func (m FunctionDeployment) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the Namespace property
func (m *FunctionDeployment) SetNamespace(val string) {
	m.Namespace = val
}

// GetNetwork returns the Network property
func (m FunctionDeployment) GetNetwork() string {
	return m.Network
}

// SetNetwork sets the Network property
func (m *FunctionDeployment) SetNetwork(val string) {
	m.Network = val
}

// GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem property
func (m FunctionDeployment) GetReadOnlyRootFilesystem() bool {
	return m.ReadOnlyRootFilesystem
}

// SetReadOnlyRootFilesystem sets the ReadOnlyRootFilesystem property
func (m *FunctionDeployment) SetReadOnlyRootFilesystem(val bool) {
	m.ReadOnlyRootFilesystem = val
}

// GetRegistryAuth returns the RegistryAuth property
func (m FunctionDeployment) GetRegistryAuth() string {
	return m.RegistryAuth
}

// SetRegistryAuth sets the RegistryAuth property
func (m *FunctionDeployment) SetRegistryAuth(val string) {
	m.RegistryAuth = val
}

// GetRequests returns the Requests property
func (m FunctionDeployment) GetRequests() *FunctionResources {
	return m.Requests
}

// SetRequests sets the Requests property
func (m *FunctionDeployment) SetRequests(val *FunctionResources) {
	m.Requests = val
}

// GetSecrets returns the Secrets property
func (m FunctionDeployment) GetSecrets() []string {
	return m.Secrets
}

// SetSecrets sets the Secrets property
func (m *FunctionDeployment) SetSecrets(val []string) {
	m.Secrets = val
}

// GetService returns the Service property
func (m FunctionDeployment) GetService() string {
	return m.Service
}

// SetService sets the Service property
func (m *FunctionDeployment) SetService(val string) {
	m.Service = val
}
