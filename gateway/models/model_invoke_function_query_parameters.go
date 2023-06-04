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

// InvokeFunctionQueryParameters is an object.
type InvokeFunctionQueryParameters struct {
	// FunctionName: Function name
	FunctionName string `json:"functionName" mapstructure:"functionName"`
}

// Validate implements basic validation for this model
func (m InvokeFunctionQueryParameters) Validate() error {
	return validation.Errors{}.Filter()
}

// GetFunctionName returns the FunctionName property
func (m InvokeFunctionQueryParameters) GetFunctionName() string {
	return m.FunctionName
}

// SetFunctionName sets the FunctionName property
func (m *InvokeFunctionQueryParameters) SetFunctionName(val string) {
	m.FunctionName = val
}
