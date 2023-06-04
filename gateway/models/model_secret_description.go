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

// SecretDescription is an object.
type SecretDescription struct {
	// Name: Name of secret
	Name string `json:"name" mapstructure:"name"`
	// Namespace: Namespace of secret
	Namespace string `json:"namespace,omitempty" mapstructure:"namespace,omitempty"`
}

// Validate implements basic validation for this model
func (m SecretDescription) Validate() error {
	return validation.Errors{}.Filter()
}

// GetName returns the Name property
func (m SecretDescription) GetName() string {
	return m.Name
}

// SetName sets the Name property
func (m *SecretDescription) SetName(val string) {
	m.Name = val
}

// GetNamespace returns the Namespace property
func (m SecretDescription) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the Namespace property
func (m *SecretDescription) SetNamespace(val string) {
	m.Namespace = val
}
