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

// GetFunctionStatusQueryParameters is an object.
type GetFunctionStatusQueryParameters struct {
	// FunctionName: Function name
	FunctionName string `json:"functionName" mapstructure:"functionName"`
	// Namespace: Namespace of the function
	Namespace string `json:"namespace,omitempty" mapstructure:"namespace,omitempty"`
}

// Validate implements basic validation for this model
func (m GetFunctionStatusQueryParameters) Validate() error {
	return validation.Errors{}.Filter()
}

// GetFunctionName returns the FunctionName property
func (m GetFunctionStatusQueryParameters) GetFunctionName() string {
	return m.FunctionName
}

// SetFunctionName sets the FunctionName property
func (m *GetFunctionStatusQueryParameters) SetFunctionName(val string) {
	m.FunctionName = val
}

// GetNamespace returns the Namespace property
func (m GetFunctionStatusQueryParameters) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the Namespace property
func (m *GetFunctionStatusQueryParameters) SetNamespace(val string) {
	m.Namespace = val
}
