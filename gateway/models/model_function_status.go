// This file is auto-generated, DO NOT EDIT.
//
// Source:
//
//	Title: OpenFaaS API Gateway
//	Version: 0.8.12
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"time"
)

// FunctionStatus is an object.
type FunctionStatus struct {
	// Annotations: A map of annotations for management, orchestration, events and build tasks
	Annotations map[string]string `json:"annotations,omitempty" mapstructure:"annotations,omitempty"`
	// AvailableReplicas: The current available amount of replicas
	AvailableReplicas float32 `json:"availableReplicas,omitempty" mapstructure:"availableReplicas,omitempty"`
	// Constraints:
	Constraints []string `json:"constraints,omitempty" mapstructure:"constraints,omitempty"`
	// CreatedAt: is the time read back from the faas backend's
	// data store for when the function or its container was created.
	CreatedAt time.Time `json:"createdAt,omitempty" mapstructure:"createdAt,omitempty"`
	// EnvProcess: Process for watchdog to fork
	EnvProcess string `json:"envProcess,omitempty" mapstructure:"envProcess,omitempty"`
	// EnvVars: environment variables for the function runtime
	EnvVars map[string]string `json:"envVars,omitempty" mapstructure:"envVars,omitempty"`
	// Image: The fully qualified docker image name of the function
	Image string `json:"image" mapstructure:"image"`
	// InvocationCount: The amount of invocations for the specified function
	InvocationCount float32 `json:"invocationCount,omitempty" mapstructure:"invocationCount,omitempty"`
	// Labels: A map of labels for making scheduling or routing decisions
	Labels map[string]string `json:"labels,omitempty" mapstructure:"labels,omitempty"`
	// Limits:
	Limits *FunctionResources `json:"limits,omitempty" mapstructure:"limits,omitempty"`
	// Name: The name of the function
	Name string `json:"name" mapstructure:"name"`
	// Namespace: The namespace of the function
	Namespace string `json:"namespace,omitempty" mapstructure:"namespace,omitempty"`
	// ReadOnlyRootFilesystem: removes write-access from the root filesystem mount-point.
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem,omitempty" mapstructure:"readOnlyRootFilesystem,omitempty"`
	// Replicas: The current minimal ammount of replicas
	Replicas float32 `json:"replicas,omitempty" mapstructure:"replicas,omitempty"`
	// Requests:
	Requests *FunctionResources `json:"requests,omitempty" mapstructure:"requests,omitempty"`
	// Secrets:
	Secrets []string `json:"secrets,omitempty" mapstructure:"secrets,omitempty"`
	// Usage:
	Usage *FunctionUsage `json:"usage,omitempty" mapstructure:"usage,omitempty"`
}

// Validate implements basic validation for this model
func (m FunctionStatus) Validate() error {
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
		"usage": validation.Validate(
			m.Usage,
		),
	}.Filter()
}

// GetAnnotations returns the Annotations property
func (m FunctionStatus) GetAnnotations() map[string]string {
	return m.Annotations
}

// SetAnnotations sets the Annotations property
func (m *FunctionStatus) SetAnnotations(val map[string]string) {
	m.Annotations = val
}

// GetAvailableReplicas returns the AvailableReplicas property
func (m FunctionStatus) GetAvailableReplicas() float32 {
	return m.AvailableReplicas
}

// SetAvailableReplicas sets the AvailableReplicas property
func (m *FunctionStatus) SetAvailableReplicas(val float32) {
	m.AvailableReplicas = val
}

// GetConstraints returns the Constraints property
func (m FunctionStatus) GetConstraints() []string {
	return m.Constraints
}

// SetConstraints sets the Constraints property
func (m *FunctionStatus) SetConstraints(val []string) {
	m.Constraints = val
}

// GetCreatedAt returns the CreatedAt property
func (m FunctionStatus) GetCreatedAt() time.Time {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt property
func (m *FunctionStatus) SetCreatedAt(val time.Time) {
	m.CreatedAt = val
}

// GetEnvProcess returns the EnvProcess property
func (m FunctionStatus) GetEnvProcess() string {
	return m.EnvProcess
}

// SetEnvProcess sets the EnvProcess property
func (m *FunctionStatus) SetEnvProcess(val string) {
	m.EnvProcess = val
}

// GetEnvVars returns the EnvVars property
func (m FunctionStatus) GetEnvVars() map[string]string {
	return m.EnvVars
}

// SetEnvVars sets the EnvVars property
func (m *FunctionStatus) SetEnvVars(val map[string]string) {
	m.EnvVars = val
}

// GetImage returns the Image property
func (m FunctionStatus) GetImage() string {
	return m.Image
}

// SetImage sets the Image property
func (m *FunctionStatus) SetImage(val string) {
	m.Image = val
}

// GetInvocationCount returns the InvocationCount property
func (m FunctionStatus) GetInvocationCount() float32 {
	return m.InvocationCount
}

// SetInvocationCount sets the InvocationCount property
func (m *FunctionStatus) SetInvocationCount(val float32) {
	m.InvocationCount = val
}

// GetLabels returns the Labels property
func (m FunctionStatus) GetLabels() map[string]string {
	return m.Labels
}

// SetLabels sets the Labels property
func (m *FunctionStatus) SetLabels(val map[string]string) {
	m.Labels = val
}

// GetLimits returns the Limits property
func (m FunctionStatus) GetLimits() *FunctionResources {
	return m.Limits
}

// SetLimits sets the Limits property
func (m *FunctionStatus) SetLimits(val *FunctionResources) {
	m.Limits = val
}

// GetName returns the Name property
func (m FunctionStatus) GetName() string {
	return m.Name
}

// SetName sets the Name property
func (m *FunctionStatus) SetName(val string) {
	m.Name = val
}

// GetNamespace returns the Namespace property
func (m FunctionStatus) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the Namespace property
func (m *FunctionStatus) SetNamespace(val string) {
	m.Namespace = val
}

// GetReadOnlyRootFilesystem returns the ReadOnlyRootFilesystem property
func (m FunctionStatus) GetReadOnlyRootFilesystem() bool {
	return m.ReadOnlyRootFilesystem
}

// SetReadOnlyRootFilesystem sets the ReadOnlyRootFilesystem property
func (m *FunctionStatus) SetReadOnlyRootFilesystem(val bool) {
	m.ReadOnlyRootFilesystem = val
}

// GetReplicas returns the Replicas property
func (m FunctionStatus) GetReplicas() float32 {
	return m.Replicas
}

// SetReplicas sets the Replicas property
func (m *FunctionStatus) SetReplicas(val float32) {
	m.Replicas = val
}

// GetRequests returns the Requests property
func (m FunctionStatus) GetRequests() *FunctionResources {
	return m.Requests
}

// SetRequests sets the Requests property
func (m *FunctionStatus) SetRequests(val *FunctionResources) {
	m.Requests = val
}

// GetSecrets returns the Secrets property
func (m FunctionStatus) GetSecrets() []string {
	return m.Secrets
}

// SetSecrets sets the Secrets property
func (m *FunctionStatus) SetSecrets(val []string) {
	m.Secrets = val
}

// GetUsage returns the Usage property
func (m FunctionStatus) GetUsage() *FunctionUsage {
	return m.Usage
}

// SetUsage sets the Usage property
func (m *FunctionStatus) SetUsage(val *FunctionUsage) {
	m.Usage = val
}
