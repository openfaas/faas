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

// ScaleServiceRequest is an object.
type ScaleServiceRequest struct {
	// Replicas: Number of replicas to scale to
	Replicas int64 `json:"replicas" mapstructure:"replicas"`
	// ServiceName: Name of deployed function
	ServiceName string `json:"serviceName" mapstructure:"serviceName"`
}

// Validate implements basic validation for this model
func (m ScaleServiceRequest) Validate() error {
	return validation.Errors{
		"replicas": validation.Validate(
			m.Replicas, validation.Min(int64(0)),
		),
	}.Filter()
}

// GetReplicas returns the Replicas property
func (m ScaleServiceRequest) GetReplicas() int64 {
	return m.Replicas
}

// SetReplicas sets the Replicas property
func (m *ScaleServiceRequest) SetReplicas(val int64) {
	m.Replicas = val
}

// GetServiceName returns the ServiceName property
func (m ScaleServiceRequest) GetServiceName() string {
	return m.ServiceName
}

// SetServiceName sets the ServiceName property
func (m *ScaleServiceRequest) SetServiceName(val string) {
	m.ServiceName = val
}
