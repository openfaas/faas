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

// DeleteFunctionRequest is an object.
type DeleteFunctionRequest struct {
	// FunctionName: Name of deployed function
	FunctionName string `json:"functionName" mapstructure:"functionName"`
}

// Validate implements basic validation for this model
func (m DeleteFunctionRequest) Validate() error {
	return validation.Errors{}.Filter()
}

// GetFunctionName returns the FunctionName property
func (m DeleteFunctionRequest) GetFunctionName() string {
	return m.FunctionName
}

// SetFunctionName sets the FunctionName property
func (m *DeleteFunctionRequest) SetFunctionName(val string) {
	m.FunctionName = val
}
