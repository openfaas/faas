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

// InvokeAsyncQueryParameters is an object.
type InvokeAsyncQueryParameters struct {
	// FunctionName: Function name
	FunctionName string `json:"functionName" mapstructure:"functionName"`
}

// Validate implements basic validation for this model
func (m InvokeAsyncQueryParameters) Validate() error {
	return validation.Errors{}.Filter()
}

// GetFunctionName returns the FunctionName property
func (m InvokeAsyncQueryParameters) GetFunctionName() string {
	return m.FunctionName
}

// SetFunctionName sets the FunctionName property
func (m *InvokeAsyncQueryParameters) SetFunctionName(val string) {
	m.FunctionName = val
}
