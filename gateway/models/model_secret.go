// This file is auto-generated, DO NOT EDIT.
//
// Source:
//
//	Title: OpenFaaS API Gateway
//	Version: 0.8.12
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Secret is an object.
type Secret struct {
	// Name: Name of secret
	Name string `json:"name" mapstructure:"name"`
	// Namespace: Namespace of secret
	Namespace string `json:"namespace,omitempty" mapstructure:"namespace,omitempty"`
	// RawValue: Value of secret in base64.
	//
	// This can be used to provide raw binary data when the `value` field is omitted.
	RawValue string `json:"rawValue,omitempty" mapstructure:"rawValue,omitempty"`
	// Value: Value of secret in plain-text
	Value string `json:"value,omitempty" mapstructure:"value,omitempty"`
}

// Validate implements basic validation for this model
func (m Secret) Validate() error {
	return validation.Errors{
		"rawValue": validation.Validate(
			m.RawValue, is.Base64,
		),
	}.Filter()
}

// GetName returns the Name property
func (m Secret) GetName() string {
	return m.Name
}

// SetName sets the Name property
func (m *Secret) SetName(val string) {
	m.Name = val
}

// GetNamespace returns the Namespace property
func (m Secret) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the Namespace property
func (m *Secret) SetNamespace(val string) {
	m.Namespace = val
}

// GetRawValue returns the RawValue property
func (m Secret) GetRawValue() string {
	return m.RawValue
}

// SetRawValue sets the RawValue property
func (m *Secret) SetRawValue(val string) {
	m.RawValue = val
}

// GetValue returns the Value property
func (m Secret) GetValue() string {
	return m.Value
}

// SetValue sets the Value property
func (m *Secret) SetValue(val string) {
	m.Value = val
}
