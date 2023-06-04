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

// SecretValues is an object.
type SecretValues struct {
	// RawValue: Value of secret in base64.
	//
	// This can be used to provide raw binary data when the `value` field is omitted.
	RawValue string `json:"rawValue,omitempty" mapstructure:"rawValue,omitempty"`
	// Value: Value of secret in plain-text
	Value string `json:"value,omitempty" mapstructure:"value,omitempty"`
}

// Validate implements basic validation for this model
func (m SecretValues) Validate() error {
	return validation.Errors{
		"rawValue": validation.Validate(
			m.RawValue, is.Base64,
		),
	}.Filter()
}

// GetRawValue returns the RawValue property
func (m SecretValues) GetRawValue() string {
	return m.RawValue
}

// SetRawValue sets the RawValue property
func (m *SecretValues) SetRawValue(val string) {
	m.RawValue = val
}

// GetValue returns the Value property
func (m SecretValues) GetValue() string {
	return m.Value
}

// SetValue sets the Value property
func (m *SecretValues) SetValue(val string) {
	m.Value = val
}
